using System;
using System.Threading.Tasks;
using Podsync.Services.Rss.Contracts;
using Podsync.Services.Storage;

namespace Podsync.Services.Rss
{
    public interface IFeedService
    {
        Task<string> Create(FeedMetadata metadata);

        Task<Feed> Get(string id);

        Task<string> Get(string id, Action<string, Feed> fixup);
    }
}