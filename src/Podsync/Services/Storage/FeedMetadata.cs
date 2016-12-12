using Podsync.Services.Links;
using Podsync.Services.Resolver;

namespace Podsync.Services.Storage
{
    public struct FeedMetadata
    {
        public Provider Provider { get; set; }

        public LinkType LinkType { get; set; }

        public string Id { get; set; }

        public ResolveType Quality { get; set; }

        public int PageSize { get; set; }

        public override string ToString() => $"{Provider} ({LinkType}) {Id}";
    }
}