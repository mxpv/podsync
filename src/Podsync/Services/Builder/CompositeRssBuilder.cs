using System;
using System.Threading.Tasks;
using Podsync.Services.Feed;
using Podsync.Services.Links;
using Podsync.Services.Storage;
using Shared;

namespace Podsync.Services.Builder
{
    // ReSharper disable once ClassNeverInstantiated.Global
    public class CompositeRssBuilder : RssBuilderBase
    {
        private readonly YouTubeRssBuilder _youTubeBuilder;
        private readonly VimeoRssBuilder _vimeoBuilder;

        public CompositeRssBuilder(IServiceProvider serviceProvider, IStorageService storageService) : base(storageService)
        {
            _youTubeBuilder = serviceProvider.CreateInstance<YouTubeRssBuilder>();
            _vimeoBuilder = serviceProvider.CreateInstance<VimeoRssBuilder>();
        }

        public override Provider Provider
        {
            get { throw new NotSupportedException(); }
        }

        public override Task<Rss> Query(Uri baseUrl, string feedId, FeedMetadata feed)
        {
            if (feed.Provider == Provider.YouTube)
            {
                return _youTubeBuilder.Query(baseUrl, feedId, feed);
            }

            if (feed.Provider == Provider.Vimeo)
            {
                return _vimeoBuilder.Query(baseUrl, feedId, feed);
            }

            throw new NotSupportedException("Not supported provider");
        }
    }
}