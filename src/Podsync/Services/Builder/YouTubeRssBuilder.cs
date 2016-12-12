using System;
using System.Linq;
using System.Threading.Tasks;
using Podsync.Services.Feed;
using Podsync.Services.Links;
using Podsync.Services.Resolver;
using Podsync.Services.Storage;
using Podsync.Services.Videos.YouTube;
using Channel = Podsync.Services.Feed.Channel;
using Video = Podsync.Services.Videos.YouTube.Video;

namespace Podsync.Services.Builder
{
    public class YouTubeRssBuilder : RssBuilderBase
    {
        private readonly ILinkService _linkService;
        private readonly IYouTubeClient _youTube;

        public YouTubeRssBuilder(ILinkService linkService, IYouTubeClient youTube, IStorageService storageService) : base(storageService)
        {
            _linkService = linkService;
            _youTube = youTube;
        }

        public override Provider Provider { get; } = Provider.YouTube;

        public override async Task<Rss> Query(FeedMetadata metadata)
        {
            if (metadata.Provider != Provider.YouTube)
            {
                throw new ArgumentException("Invalid provider");
            }

            var linkType = metadata.LinkType;

            Channel channel;
            if (linkType == LinkType.Channel)
            {
                channel = await GetChannel(new ChannelQuery { ChannelId = metadata.Id });
            }
            else if (linkType == LinkType.User)
            {
                channel = await GetChannel(new ChannelQuery { Username = metadata.Id });
            }
            else if (linkType == LinkType.Playlist)
            {
                channel = await GetPlaylist(metadata.Id);
            }
            else
            {
                throw new NotSupportedException("URL type is not supported");
            }

            // Get video ids from this playlist
            var ids = await _youTube.GetPlaylistItemIds(new PlaylistItemsQuery { PlaylistId = channel.Guid });

            // Get video descriptions
            var videos = await _youTube.GetVideos(new VideoQuery { Id = string.Join(",", ids) });

            channel.Items = videos.Select(x => MakeItem(x, metadata));

            var rss = new Rss
            {
                Channels = new[] { channel }
            };

            return rss;
        }

        private async Task<Channel> GetChannel(ChannelQuery query)
        {
            var list = await _youTube.GetChannels(query);
            var item = list.Single();

            var channel = MakeChannel(item);
            channel.Guid = item.PlaylistId;

            return channel;
        }

        private async Task<Channel> GetPlaylist(string playlistId)
        {
            var list = await _youTube.GetPlaylists(new PlaylistQuery { PlaylistId = playlistId });
            var item = list.Single();

            var channel = MakeChannel(item);
            channel.Guid = item.PlaylistId;

            return channel;
        }

        private static Channel MakeChannel(YouTubeItem item)
        {
            return new Channel
            {
                Title = item.Title,
                Description = item.Description,
                Link = item.Link,
                LastBuildDate = DateTime.Now,
                PubDate = item.PublishedAt,
                Image = item.Thumbnail,
                Thumbnail = item.Thumbnail
            };
        }

        private Item MakeItem(Video video, FeedMetadata feed)
        {
            return new Item
            {
                Title = video.Title,
                Description = video.Description,
                PubDate = video.PublishedAt,
                Link = video.Link,
                Duration = video.Duration,
                Content = new MediaContent
                {
                    Length = video.Size,
                    MediaType = SelectMediaType(feed.Quality),
                    Url = _linkService.Download(feed.Id, video.VideoId)
                }
            };
        }

        private static string SelectMediaType(ResolveType resolveType)
        {
            return resolveType == ResolveType.VideoHigh || resolveType == ResolveType.VideoLow
                ? "video/mp4"
                : "audio/webm";
        }
    }
}