using System.Threading.Tasks;
using Podsync.Services.Storage;

namespace Podsync.Services.Rss
{
    public interface IFeedService
    {
        Task<string> Create(FeedMetadata metadata);

        Task<Contracts.Feed> Get(string id);
    }
}