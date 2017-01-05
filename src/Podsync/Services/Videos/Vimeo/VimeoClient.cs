using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Threading.Tasks;
using Microsoft.Extensions.Options;
using Newtonsoft.Json.Linq;

namespace Podsync.Services.Videos.Vimeo
{
    // ReSharper disable once ClassNeverInstantiated.Global
    public sealed class VimeoClient : IVimeoClient, IDisposable
    {
        private const int MaxPageSize = 100;

        private readonly HttpClient _client = new HttpClient();

        public VimeoClient(IOptions<PodsyncConfiguration> configuration)
        {
            _client.BaseAddress = new Uri("https://api.vimeo.com/");
            _client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", configuration.Value.VimeoApiKey);
        }

        public Task<Group> Group(string id)
        {
            return QueryGroup($"groups/{id}");
        }

        public Task<Group> Channel(string id)
        {
            return QueryGroup($"channels/{id}");
        }

        public async Task<User> User(string id)
        {
            dynamic json = await QueryApi($"users/{id}");

            return new User
            {
                Name = json.name,
                Bio = json.bio,
                Link = new Uri(json.link.ToString()),
                Thumbnail = new Uri(json.pictures.sizes[0].link.ToString()),
                CreatedAt = DateTime.Parse(json.created_time.ToString()),
            };
        }

        public Task<IEnumerable<Video>> GroupVideos(string id, int count)
        {
            return QueryVideos($"groups/{id}/videos", count);
        }

        public Task<IEnumerable<Video>> UserVideos(string id, int count)
        {
            return QueryVideos($"users/{id}/videos", count);
        }

        public Task<IEnumerable<Video>> ChannelVideos(string id, int count)
        {
            return QueryVideos($"channels/{id}/videos", count);
        }

        public void Dispose()
        {
            _client.Dispose();
        }

        private async Task<IEnumerable<Video>> QueryVideos(string path, int count)
        {
            if (count <= 0)
            {
                throw new ArgumentException("Invalid item count", nameof(count));
            }

            var collection = new List<Video>(count);
            var pageIndex = 1;

            while (count > 0)
            {
                var pageSize = Math.Min(count, MaxPageSize);

                await GetPage(path, pageIndex, pageSize, collection);

                count -= pageSize;
                pageIndex++;
            }

            return collection;
        }

        private async Task GetPage(string path, int pageIndex, int pageSize, List<Video> output)
        {
            dynamic resp = await QueryApi($"{path}?per_page={pageSize}&page={pageIndex}");

            foreach (dynamic v in resp.data)
            {
                // Approximated file size
                var size = Convert.ToInt64(
                    v.width.ToObject<long>() *
                    v.height.ToObject<long>() *
                    v.duration.ToObject<long>() *
                    0.38848958333);

                var video = new Video
                {
                    Title = v.name,
                    Description = v.description,
                    Link = new Uri(v.link?.ToString()),
                    Thumbnail = new Uri(v.pictures?.sizes[0]?.link?.ToString()),
                    CreatedAt = DateTime.Parse(v.created_time?.ToString()),
                    Duration = TimeSpan.FromSeconds(v.duration?.ToObject<int>()),
                    Size = size
                };

                output.Add(video);
            }
        }

        private async Task<Group> QueryGroup(string path)
        {
            dynamic json = await QueryApi(path);

            return new Group
            {
                Name = json.name,
                Description = json.description,
                Link = new Uri(json.link?.ToString()),
                Thumbnail = new Uri(json.pictures?.sizes[0]?.link?.ToString()),
                CreatedAt = DateTime.Parse(json.created_time?.ToString()),
                Author = json.user.name,
            };
        }

        private async Task<JObject> QueryApi(string path)
        {
            var json = await _client.GetStringAsync(path);
            return JObject.Parse(json);
        }
    }
}