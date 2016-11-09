using System;
using System.Threading.Tasks;

namespace Podsync.Services.Storage
{
    public interface IStorageService : IDisposable
    {
        Task<TimeSpan> Ping();

        Task<string> Save(FeedMetadata metadata);

        Task<FeedMetadata> Load(string key);
    }
}