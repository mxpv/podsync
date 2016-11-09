using Podsync.Services.Links;

namespace Podsync.Services.Storage
{
    public struct FeedMetadata
    {
        public Provider Provider { get; set; }

        public LinkType LinkType { get; set; }

        public string Id { get; set; }
    }
}