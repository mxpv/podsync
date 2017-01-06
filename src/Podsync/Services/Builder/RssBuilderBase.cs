using System;
using System.Threading.Tasks;
using Podsync.Services.Feed;
using Podsync.Services.Links;
using Podsync.Services.Storage;

namespace Podsync.Services.Builder
{
    public abstract class RssBuilderBase : IRssBuilder
    {
        private readonly IStorageService _storageService;

        protected RssBuilderBase(IStorageService storageService)
        {
            _storageService = storageService;
        }

        public abstract Provider Provider { get; }

        public async Task<Rss> Query(string feedId)
        {
            var metadata = await _storageService.Load(feedId);

            return await Query(metadata);
        }

        public abstract Task<Rss> Query(FeedMetadata metadata);
    }
}