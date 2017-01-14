using System;
using System.Linq;
using System.Threading.Tasks;
using Podsync.Services.Links;
using Podsync.Services.Resolver;
using Podsync.Services.Rss.Contracts;
using Podsync.Services.Storage;
using Podsync.Services.Videos.YouTube;
using Channel = Podsync.Services.Rss.Contracts.Channel;
using Video = Podsync.Services.Videos.YouTube.Video;

namespace Podsync.Services.Rss.Builders
{
    public class YouTubeRssBuilder : RssBuilderBase
    {
        private readonly IYouTubeClient _youTube;

        public YouTubeRssBuilder(IYouTubeClient youTube, IStorageService storageService) : base(storageService)
        {
            _youTube = youTube;
        }

        public override Provider Provider { get; } = Provider.YouTube;

        public override async Task<Feed> Query(FeedMetadata metadata)
        {
            if (metadata.Provider != Provider.YouTube)
            {
                throw new ArgumentException("Invalid provider");
            }

            if (metadata.PageSize == 0)
            {
                metadata.PageSize = Constants.DefaultPageSize;
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
            var ids = await _youTube.GetPlaylistItemIds(new PlaylistItemsQuery { PlaylistId = channel.Guid, Count = metadata.PageSize });

            // Get video descriptions
            var videos = await _youTube.GetVideos(new VideoQuery { Ids = ids });

            channel.Items = videos.Select(youtubeVideo => MakeItem(youtubeVideo, metadata)).ToArray();

            var rss = new Feed
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
                PubDate = item.PublishedAt,
                Image = item.Thumbnail,
                Thumbnail = item.Thumbnail,
            };
        }

        private Item MakeItem(Video video, FeedMetadata feed)
        {
            string contentType = GetContentType(feed.Quality);

            return new Item
            {
                Id = video.VideoId,
                Title = video.Title,
                Description = video.Description,
                PubDate = video.PublishedAt,
                Link = video.Link,
                Duration = video.Duration,
                FileSize = video.Size,
                ContentType = contentType
            };
        }

        private static string GetContentType(ResolveFormat format)
        {
            if (format == ResolveFormat.VideoHigh || format == ResolveFormat.VideoLow)
            {
                return "video/mp4";
            }

            if (format == ResolveFormat.AudioHigh || format == ResolveFormat.AudioLow)
            {
                return "audio/mp4";
            }

            throw new ArgumentException("Unsupported resolve type");
        }
    }
}