using System;
using System.Threading.Tasks;
using Podsync.Helpers;
using Podsync.Services.Links;
using Podsync.Services.Rss.Contracts;
using Podsync.Services.Storage;

namespace Podsync.Services.Rss
{
    public class FeedService : IFeedService
    {
        private readonly IStorageService _storageService;
        private readonly IRssBuilder _rssBuilder;

        public FeedService(IStorageService storageService, IRssBuilder rssBuilder)
        {
            _storageService = storageService;
            _rssBuilder = rssBuilder;
        }

        public Task<string> Create(FeedMetadata metadata)
        {
            if (metadata.Provider != Provider.YouTube && metadata.Quality.IsAudio())
            {
                throw new ArgumentException("Only YouTube supports audio feeds");
            }

            return _storageService.Save(metadata);
        }

        public Task<Feed> Get(string id)
        {
            return _rssBuilder.Query(id);
        }

        public async Task<string> Get(string id, Action<string, Feed> fixup)
        {
            var serializedFeed = await _storageService.GetCached(Constants.Cache.FeedsPrefix, id);

            if (string.IsNullOrEmpty(serializedFeed))
            {
                var feed = await Get(id);

                // Fix download links
                fixup(id, feed);

                // Add to cache
                serializedFeed = feed.ToString();
                await _storageService.Cache(Constants.Cache.FeedsPrefix, id, serializedFeed, TimeSpan.FromMinutes(3));
            }

            return serializedFeed;
        }
    }
}