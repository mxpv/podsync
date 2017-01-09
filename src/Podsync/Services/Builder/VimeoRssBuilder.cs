using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Podsync.Services.Feed;
using Podsync.Services.Links;
using Podsync.Services.Storage;
using Podsync.Services.Videos.Vimeo;

namespace Podsync.Services.Builder
{
    // ReSharper disable once ClassNeverInstantiated.Global
    public class VimeoRssBuilder : RssBuilderBase
    {
        private readonly IVimeoClient _client;

        public VimeoRssBuilder(IStorageService storageService, IVimeoClient client) : base(storageService)
        {
            _client = client;
        }

        public override Provider Provider { get; } = Provider.Vimeo;

        public override async Task<Rss> Query(FeedMetadata metadata)
        {
            var linkType = metadata.LinkType;

            var id = metadata.Id;

            var pageSize = metadata.PageSize;
            if (pageSize == 0)
            {
                pageSize = Constants.DefaultPageSize;
            }

            Channel channel;
            if (linkType == LinkType.Channel)
            {
                channel = CreateChannel(await _client.Channel(id));
                channel.Items = CreateItems(await _client.ChannelVideos(id, pageSize));
            }
            else if (linkType == LinkType.Group)
            {
                channel = CreateChannel(await _client.Group(id));
                channel.Items = CreateItems(await _client.GroupVideos(id, pageSize));
            }
            else if (linkType == LinkType.User)
            {
                channel = CreateChannel(await _client.User(id));
                channel.Items = CreateItems(await _client.UserVideos(id, pageSize));
            }
            else
            {
                throw new NotSupportedException("URL type is not supported");
            }

            var rss = new Rss
            {
                Channels = new[] { channel }
            };

            return rss;
        }

        private static Channel CreateChannel(Group group)
        {
            return new Channel
            {
                Title = group.Name,
                Description = group.Description,
                Link = group.Link,
                PubDate = group.CreatedAt,
                Image = group.Thumbnail,
                Thumbnail = group.Thumbnail,
                Guid = group.Link.ToString()
            };
        }

        private static Channel CreateChannel(User user)
        {
            return new Channel
            {
                Title = user.Name,
                Description = user.Bio,
                Link = user.Link,
                PubDate = user.CreatedAt,
                Image = user.Thumbnail,
                Thumbnail = user.Thumbnail,
                Guid = user.Link.ToString()
            };
        }

        private static Item CreateItem(Video video)
        {
            return new Item
            {
                Id = video.Id,
                Title = video.Title,
                Description = video.Description,
                PubDate = video.CreatedAt,
                Link = video.Link,
                Duration = video.Duration,
                FileSize = video.Size,
                ContentType = "video/mp4",
                Author = video.Author
            };
        }

        private static Item[] CreateItems(IEnumerable<Video> videos)
        {
            return videos.Select(CreateItem).ToArray();
        }
    }
}