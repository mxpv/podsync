using System.Threading.Tasks;
using Podsync.Services.Links;
using Podsync.Services.Storage;

namespace Podsync.Services.Rss.Builders
{
    public abstract class RssBuilderBase : IRssBuilder
    {
        private readonly IStorageService _storageService;

        protected RssBuilderBase(IStorageService storageService)
        {
            _storageService = storageService;
        }

        public abstract Provider Provider { get; }

        public async Task<Contracts.Feed> Query(string feedId)
        {
            var metadata = await _storageService.Load(feedId);

            return await Query(metadata);
        }

        public abstract Task<Contracts.Feed> Query(FeedMetadata metadata);
    }
}