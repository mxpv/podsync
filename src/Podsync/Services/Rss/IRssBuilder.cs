using System.Threading.Tasks;
using Podsync.Services.Links;
using Podsync.Services.Storage;

namespace Podsync.Services.Rss
{
    public interface IRssBuilder
    {
        Provider Provider { get; }

        Task<Feed.Rss> Query(string feedId);

        Task<Feed.Rss> Query(FeedMetadata metadata);
    }
}