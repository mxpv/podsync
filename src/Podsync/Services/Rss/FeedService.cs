using System;
using System.Threading.Tasks;
using Podsync.Helpers;
using Podsync.Services.Links;
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

            metadata.PageSize = Constants.DefaultPageSize;

            return _storageService.Save(metadata);
        }

        public Task<Contracts.Feed> Get(string id)
        {
            return _rssBuilder.Query(id);
        }
    }
}