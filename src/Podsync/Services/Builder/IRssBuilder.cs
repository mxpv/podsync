using System.Threading.Tasks;
using Podsync.Services.Feed;
using Podsync.Services.Storage;

namespace Podsync.Services.Builder
{
    public interface IRssBuilder
    {
        Task<Rss> Query(string feedId);

        Task<Rss> Query(string feedId, FeedMetadata feed);
    }
}