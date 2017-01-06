using System;
using System.Threading.Tasks;
using Podsync.Services.Feed;
using Podsync.Services.Links;
using Podsync.Services.Storage;

namespace Podsync.Services.Builder
{
    public interface IRssBuilder
    {
        Provider Provider { get; }

        Task<Rss> Query(Uri baseUrl, string feedId);

        Task<Rss> Query(Uri baseUrl, string feedId, FeedMetadata metadata);
    }
}