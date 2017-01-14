using System.Threading.Tasks;
using Podsync.Services.Links;
using Podsync.Services.Rss.Contracts;
using Podsync.Services.Storage;

namespace Podsync.Services.Rss
{
    public interface IRssBuilder
    {
        Provider Provider { get; }

        Task<Feed> Query(string feedId);

        Task<Feed> Query(FeedMetadata metadata);
    }
}