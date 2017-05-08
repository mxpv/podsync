using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using System.Xml;
using Google.Apis.Services;
using Google.Apis.YouTube.v3;
using Google.Apis.YouTube.v3.Data;
using Microsoft.Extensions.Options;
using Podsync.Helpers;
using Podsync.Services.Links;

namespace Podsync.Services.Videos.YouTube
{
    public sealed class YouTubeClient : IYouTubeClient, IDisposable
    {
        private const int MaxResults = 50;

        private readonly ILinkService _linkService;
        private readonly YouTubeService _youtube;

        public YouTubeClient(ILinkService linkService, IOptions<PodsyncConfiguration> configuration)
        {
            _linkService = linkService;
            _youtube = new YouTubeService(new BaseClientService.Initializer
            {
                ApplicationName = "Podsync",
                ApiKey = configuration.Value.YouTubeApiKey
            });
        }

        public Task<ICollection<Channel>> GetChannels(ChannelQuery query)
        {
            var request = _youtube.Channels.List("id,snippet,contentDetails");

            request.Id = query.ChannelId;
            request.ForUsername = query.Username;

            return AggregatePages<Channel>(query.Count, async (list, token, pageSize) =>
            {
                request.MaxResults = pageSize;
                request.PageToken = token;

                var response = await request.ExecuteAsync();
                response.Items.Select(ConvertChannel).AddTo(list);

                return response.NextPageToken;
            });
        }

        public Task<ICollection<Playlist>> GetPlaylists(PlaylistQuery query)
        {
            var request = _youtube.Playlists.List("id,snippet");

            request.Id = query.PlaylistId;
            request.ChannelId = query.ChannelId;

            return AggregatePages<Playlist>(query.Count, async (list, token, pageSize) =>
            {
                request.MaxResults = pageSize;
                request.PageToken = token;

                var response = await request.ExecuteAsync();
                response.Items.Select(ConvertPlaylist).AddTo(list);

                return response.NextPageToken;
            });
        }

        public Task<ICollection<Video>> GetVideos(VideoQuery query)
        {
            var request = _youtube.Videos.List("id,snippet,contentDetails");

            var totalCount = query.Ids.Count;
            var pageIndex = 0;

            return AggregatePages<Video>(totalCount, async (list, token, pageSize) =>
            {
                request.Id = string.Join(",", query.Ids.Skip(pageIndex * MaxResults).Take(pageSize));
                if (string.IsNullOrEmpty(request.Id))
                {
                    return null;
                }

                pageIndex++;

                var response = await request.ExecuteAsync();
                response.Items.Select(ConvertVideo).AddTo(list);

                return list.Count == totalCount ? null : true.ToString();
            });
        }

        public Task<ICollection<Video>> GetPlaylistItems(PlaylistItemsQuery query)
        {
            var request = _youtube.PlaylistItems.List("id,snippet");

            request.Id = query.Id;
            request.PlaylistId = query.PlaylistId;
            request.VideoId = query.VideoId;

            return AggregatePages<Video>(query.Count, async (list, token, pageSize) =>
            {
                request.MaxResults = pageSize;
                request.PageToken = token;

                var response = await request.ExecuteAsync();
                response.Items.Select(ConvertPlaylistItem).AddTo(list);

                return response.NextPageToken;
            });
        }

        public Task<ICollection<string>> GetPlaylistItemIds(PlaylistItemsQuery query)
        {
            var request = _youtube.PlaylistItems.List("id,snippet");

            request.PlaylistId = query.PlaylistId;

            return AggregatePages<string>(query.Count, async (list, token, pageSize) =>
            {
                request.MaxResults = pageSize;
                request.PageToken = token;

                var response = await request.ExecuteAsync();
                response.Items.Select(x => x.Snippet.ResourceId.VideoId).AddTo(list);

                return response.NextPageToken;
            });
        }

        public void Dispose()
        {
            _youtube.Dispose();
        }

        private static long GetVideoSize(string definition, TimeSpan duration)
        {
            // Video size information requires 1 additional call for each video (1 feed = 50 videos = 50 calls),
            // which is too expensive, so get approximated size depending on duration and definition params

            var totalSeconds = (long)duration.TotalSeconds;

            const long hdBytesPerSecond = 350000;
            const long ldBytesPerSecond = 100000;

            return totalSeconds * (definition == "hd" ? hdBytesPerSecond : ldBytesPerSecond);
        }

        private static async Task<ICollection<T>> AggregatePages<T>(int? totalCount, Func<List<T>, string, int, Task<string>> fetchPage)
        {
            var count = totalCount ?? MaxResults;
            if (count <= 0)
            {
                throw new ArgumentException("Invalid page count");
            }

            var token = string.Empty;

            var list = new List<T>(count);
            while (token != null && list.Count < count)
            {
                var pageSize = Math.Min(MaxResults, count - list.Count);
                token = await fetchPage(list, token, pageSize);
            }

            return list;
        }

        private Video ConvertVideo(Google.Apis.YouTube.v3.Data.Video item)
        {
            var snippet = item.Snippet;
            var details = item.ContentDetails;

            var link = _linkService.Make(new LinkInfo
            {
                Id = item.Id,
                LinkType = LinkType.Video,
                Provider = Provider.YouTube
            });

            var duration = XmlConvert.ToTimeSpan(details.Duration);
            var size = GetVideoSize(details.Definition, duration);

            return new Video
            {
                VideoId = item.Id,
                ChannelId = snippet.ChannelId,
                Title = snippet.Title,
                Description = snippet.Description,
                PublishedAt = snippet.PublishedAt ?? DateTime.MinValue,
                Link = link,
                Duration = duration,
                Size = size,
            };
        }

        private Video ConvertPlaylistItem(PlaylistItem item)
        {
            var snippet = item.Snippet;

            var link = _linkService.Make(new LinkInfo
            {
                Id = item.Id,
                LinkType = LinkType.Video,
                Provider = Provider.YouTube
            });

            return new Video
            {
                VideoId = snippet.ResourceId.VideoId,
                ChannelId = snippet.ChannelId,
                PlaylistId = snippet.PlaylistId,
                Link = link,
                Title = snippet.Title,
                Description = snippet.Description,
                PublishedAt = snippet.PublishedAt ?? DateTime.MinValue,
                Thumbnail = new Uri(snippet.Thumbnails.Default__.Url)
            };
        }

        private Playlist ConvertPlaylist(Google.Apis.YouTube.v3.Data.Playlist item)
        {
            var link = _linkService.Make(new LinkInfo
            {
                Id = item.Id,
                LinkType = LinkType.Playlist,
                Provider = Provider.YouTube
            });

            var snippet = item.Snippet;
            
            return new Playlist
            {
                PlaylistId = item.Id,
                ChannelId = snippet.ChannelId,
                Link = link,
                Title = $"{snippet.ChannelTitle}: {snippet.Title}",
                Description = snippet.Description,
                PublishedAt = snippet.PublishedAt ?? DateTime.MinValue,
                Thumbnail = new Uri(snippet.Thumbnails.Default__.Url)
            };
        }

        private Channel ConvertChannel(Google.Apis.YouTube.v3.Data.Channel item)
        {
            var link = _linkService.Make(new LinkInfo
            {
                Id = item.Id,
                LinkType = LinkType.Channel,
                Provider = Provider.YouTube
            });

            var snippet = item.Snippet;

            return new Channel
            {
                ChannelId = item.Id,
                PlaylistId = item.ContentDetails.RelatedPlaylists.Uploads,
                Link = link,
                Title = snippet.Title,
                Description = snippet.Description,
                PublishedAt = snippet.PublishedAt ?? DateTime.MinValue,
                Thumbnail = new Uri(snippet.Thumbnails.Default__.Url)
            };
        }
    }
}