using System;
using System.Collections.Generic;
using System.Linq;
using Microsoft.AspNetCore.WebUtilities;
using Microsoft.Extensions.Options;
using Shared;

namespace Podsync.Services.Links
{
    public class LinkService : ILinkService
    {
        private static readonly IDictionary<Provider, IDictionary<LinkType, string>> LinkFormats = new Dictionary<Provider, IDictionary<LinkType, string>>
        {
            [Provider.YouTube] = new Dictionary<LinkType, string>
            {
                [LinkType.Video] = "https://youtube.com/watch?v={0}",
                [LinkType.Channel] = "https://youtube.com/channel/{0}",
                [LinkType.Playlist] = "https://youtube.com/playlist?list={0}",
                [LinkType.Info] = "https://youtube.com/get_video_info?video_id={0}"
            },

            [Provider.Vimeo] = new Dictionary<LinkType, string>
            {
                [LinkType.Category] = "https://vimeo.com/categories/{0}",
                [LinkType.Channel] = "https://vimeo.com/channels/{0}",
                [LinkType.Group] = "https://vimeo.com/groups/{0}",
                [LinkType.User] = "https://vimeo.com/{0}",
                [LinkType.Info] = "https://player.vimeo.com/video/{0}/config"
            }
        };

        private readonly Uri _baseUrl;

        public LinkService(IOptions<PodsyncConfiguration> configuration)
        {
            _baseUrl = new Uri(configuration.Value.BaseUrl ?? "http://localhost");
        }

        public LinkInfo Parse(Uri link)
        {
            if (link == null)
            {
                throw new ArgumentNullException(nameof(link), "Link can't be null");
            }

            var provider = Provider.Unknown;
            var linkType = LinkType.Unknown;

            var id = string.Empty;

            var segments = link.Segments
                .Select(x => x.TrimEnd('/'))
                .Where(x => !string.IsNullOrWhiteSpace(x))
                .ToArray();

            var host = link.Host.ToLowerInvariant().TrimStart("www.");

            if (host == "youtu.be")
            {
                provider = Provider.YouTube;

                if (segments.Length == 1)
                {
                    // https://youtu.be/AAAAAAAAA01
                    // https://www.youtu.be/AAAAAAAAA08

                    linkType = LinkType.Video;
                    id = segments.Single();
                }
            }
            else if (host == "youtube.com")
            {
                provider = Provider.YouTube;

                var query = QueryHelpers.ParseQuery(link.Query);

                if (segments.Length >= 2 && segments[0] == "user")
                {
                    linkType = LinkType.User;
                    id = segments[1];
                }
                else if (segments.Length == 2)
                {
                    if (string.Equals(segments[0], "embed"))
                    {
                        linkType = LinkType.Video;

                        if (string.Equals(segments[1], "watch"))
                        {
                            // http://www.youtube.com/embed/watch?feature=player_embedded&v=AAAAAAAAA02
                            // http://www.youtube.com/embed/watch?v=AAAAAAAAA03

                            id = query["v"];
                        }
                        else if (segments[1].StartsWith("v="))
                        {
                            // http://www.youtube.com/embed/v=AAAAAAAAA04
                            
                            id = segments[1].TrimStart("v=");
                        }
                    }
                    else if (string.Equals(segments[0], "watch"))
                    {
                        // http://www.youtube.com/watch/jMeC7JFQ6811

                        linkType = LinkType.Video;
                        id = segments[1];
                    }
                    else if (string.Equals(segments[0], "v"))
                    {
                        // http://www.youtube.com/v/jMeC7JFQ6812
                        // http://www.youtube.com/v/A-AAAAAAA18?fs=1&rel=0

                        linkType = LinkType.Video;
                        id = segments[1];
                    }
                    else if (string.Equals(segments[0], "channel"))
                    {
                        // https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og

                        linkType = LinkType.Channel;
                        id = segments[1];
                    }
                }

                else if (segments.Length == 1)
                {
                    if (string.Equals(segments[0], "watch"))
                    {
                        if (query.ContainsKey("list"))
                        {
                            // https://www.youtube.com/watch?v=otm9NaT9OWU&list=PLCB9F975ECF01953C

                            linkType = LinkType.Playlist;
                            id = query["list"];
                        }
                        else
                        {
                            // http://www.youtube.com/watch?v=AAAAAAAAA06
                            // http://www.youtube.com/watch?feature=player_embedded&v=AAAAAAAAA05

                            linkType = LinkType.Video;
                            id = query["v"];
                        }
                    }
                    else if (string.Equals(segments[0], "attribution_link"))
                    {
                        // http://www.youtube.com/attribution_link?u=/watch?v=jMeC7JFQ6815&feature=share&a=9QlmP1yvjcllp0h3l0NwuA
                        // http://www.youtube.com/attribution_link?a=fF1CWYwxCQ4&u=/watch?v=jMeC7JFQ6816&feature=em-uploademail 
                        // http://www.youtube.com/attribution_link?a=fF1CWYwxCQ4&feature=em-uploademail&u=/watch?v=jMeC7JFQ6817 

                        string u = query["u"];

                        var pos = u?.IndexOf("?", StringComparison.OrdinalIgnoreCase) ?? -1;
                        if (pos != -1)
                        {
                            // ReSharper disable once PossibleNullReferenceException
                            var attrQueryParams = QueryHelpers.ParseQuery(u.Substring(pos));

                            linkType = LinkType.Video;
                            id = attrQueryParams["v"];
                        }
                    }
                    else if (string.Equals(segments[0], "playlist"))
                    {
                        // https://www.youtube.com/playlist?list=PLCB9F975ECF01953C

                        linkType = LinkType.Playlist;
                        id = query["list"];
                    }
                }
            }

            if (string.IsNullOrWhiteSpace(id) || linkType == LinkType.Unknown || provider == Provider.Unknown)
            {
                throw new ArgumentException("This provider is not supported");
            }
            
            return new LinkInfo
            {
                Id = id,
                LinkType = linkType,
                Provider = provider
            };
        }

        public Uri Make(LinkInfo info)
        {
            if (info.Id == null)
            {
                throw new ArgumentNullException(nameof(info.Id), "Id can't be empty");
            }

            try
            {
                var format = LinkFormats[info.Provider][info.LinkType];
                return new Uri(string.Format(format, info.Id));
            }
            catch (KeyNotFoundException ex)
            {
                throw new ArgumentException("Unsupported provider or link type", nameof(info), ex);
            }
        }

        public Uri Download(string feedId, string videoId)
        {
            return new Uri(_baseUrl, $"download/{feedId}/{videoId}/");
        }
    }
}