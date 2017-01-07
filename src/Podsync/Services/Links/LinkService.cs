using System;
using System.Collections.Generic;
using System.Text.RegularExpressions;
using Microsoft.AspNetCore.WebUtilities;
using Microsoft.Extensions.Primitives;

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
                [LinkType.Playlist] = "https://youtube.com/playlist?list={0}"
            },

            [Provider.Vimeo] = new Dictionary<LinkType, string>
            {
                [LinkType.Channel] = "https://vimeo.com/channels/{0}",
                [LinkType.Group] = "https://vimeo.com/groups/{0}",
                [LinkType.User] = "https://vimeo.com/user{0}",
                [LinkType.Video] = "https://vimeo.com/{0}"
            }
        };

        /*
            YouTube users, channels and playlists
            Test input:
            https://www.youtube.com/playlist?list=PLCB9F975ECF01953C
            https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og
            https://www.youtube.com/user/fxigr1
         */

        private static readonly Regex YouTubeRegex = new Regex(@"^(?:https?://)?(?:www\.)?(?:youtube.com/)(?<type>user|channel|playlist|watch)/?(?<id>\w+)?", RegexOptions.Compiled);

        /*
            Vimeo groups, channels and users
            Test input:
            https://vimeo.com/groups/109
            http://vimeo.com/groups/109
            http://www.vimeo.com/groups/109
            https://vimeo.com/groups/109/videos/
            https://vimeo.com/channels/staffpicks
            https://vimeo.com/channels/staffpicks/146224925
            https://vimeo.com/awhitelabelproduct
        */
        private static readonly Regex VimeoRegex = new Regex(@"^(?:https?://)?(?:www\.)?(?:vimeo.com/)(?<type>groups|channels)?/?(?<id>\w+)", RegexOptions.Compiled);

        public LinkInfo Parse(Uri link)
        {
            if (link == null)
            {
                throw new ArgumentNullException(nameof(link), "Link can't be null");
            }

            var provider = Provider.Unknown;
            var linkType = LinkType.Unknown;

            var id = string.Empty;

            // YouTube
            var match = YouTubeRegex.Match(link.AbsoluteUri);
            if (match.Success)
            {
                provider = Provider.YouTube;

                var type = match.Groups["type"]?.ToString();
                if (type == "user")
                {
                    // https://www.youtube.com/user/fxigr1

                    id = match.Groups["id"]?.ToString();
                    linkType = LinkType.User;
                }
                else if (type == "channel")
                {
                    // https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og

                    id = match.Groups["id"]?.ToString();
                    linkType = LinkType.Channel;
                }
                else if (type == "playlist" || type == "watch")
                {
                    // https://www.youtube.com/playlist?list=PLCB9F975ECF01953C
                    // https://www.youtube.com/watch?v=otm9NaT9OWU&list=PLCB9F975ECF01953C

                    var qs = QueryHelpers.ParseQuery(link.Query);

                    StringValues list;
                    if (qs.TryGetValue("list", out list))
                    {
                        id = list;
                    }

                    linkType = LinkType.Playlist;
                }
            }
            else
            {
                // Vimeo
                match = VimeoRegex.Match(link.AbsoluteUri);
                if (match.Success)
                {
                    provider = Provider.Vimeo;
                    id = match.Groups["id"]?.ToString();

                    var type = match.Groups["type"]?.ToString();
                    if (type == "groups")
                    {
                        // https://vimeo.com/groups/109

                        linkType = LinkType.Group;
                    }
                    else if (type == "channels")
                    {
                        // https://vimeo.com/channels/staffpicks

                        linkType = LinkType.Channel;
                    }
                    else
                    {
                        // https://vimeo.com/awhitelabelproduct

                        linkType = LinkType.User;
                    }
                }
            }

            if (string.IsNullOrWhiteSpace(id) || linkType == LinkType.Unknown || provider == Provider.Unknown)
            {
                throw new ArgumentException("Not supported provider or link format");
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

        public Uri Feed(Uri baseUrl, string feedId)
        {
            return new Uri(baseUrl, feedId);
        }
    }
}