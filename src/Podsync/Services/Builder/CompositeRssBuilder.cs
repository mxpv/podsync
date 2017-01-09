using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Reflection;
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
        private readonly IDictionary<Provider, IRssBuilder> _builders;

        public CompositeRssBuilder(IServiceProvider serviceProvider, IStorageService storageService) : base(storageService)
        {
            // Find all RSS builders (all implementations of IRssBuilder), create instances and make dictionary for fast search by Provider type
            var buildTypes = serviceProvider.FindAllImplementationsOf<IRssBuilder>(Assembly.GetEntryAssembly()).Where(x => x != typeof(CompositeRssBuilder));
            var builders = buildTypes.Select(builderType => (IRssBuilder)serviceProvider.CreateInstance(builderType)).ToDictionary(builder => builder.Provider);

            _builders = new ReadOnlyDictionary<Provider, IRssBuilder>(builders);
        }

        public override Provider Provider
        {
            get { throw new NotSupportedException(); }
        }

        public override Task<Rss> Query(FeedMetadata feed)
        {
            IRssBuilder builder;
            if (_builders.TryGetValue(feed.Provider, out builder))
            {
                return builder.Query(feed);
            }

            throw new NotSupportedException("Not supported provider");
        }
    }
}