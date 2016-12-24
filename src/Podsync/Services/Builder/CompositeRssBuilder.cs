using System;
using System.Threading.Tasks;
using Podsync.Services.Feed;
using Podsync.Services.Links;
using Podsync.Services.Storage;
using Shared;

namespace Podsync.Services.Builder
{
    public class CompositeRssBuilder : RssBuilderBase
    {
        private readonly YouTubeRssBuilder _youTubeBuilder;

        public CompositeRssBuilder(IServiceProvider serviceProvider, IStorageService storageService) : base(storageService)
        {
            _youTubeBuilder = serviceProvider.CreateInstance<YouTubeRssBuilder>();
        }

        public override Provider Provider
        {
            get { throw new NotSupportedException(); }
        }

        public override Task<Rss> Query(string feedId, FeedMetadata feed)
        {
            if (feed.Provider == Provider.YouTube)
            {
                return _youTubeBuilder.Query(feedId, feed);
            }

            throw new NotSupportedException("Not supported provider");
        }
    }
}