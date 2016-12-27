using System;
using System.Threading.Tasks;
using Podsync.Services.Feed;

namespace Podsync.Services.Builder
{
    public interface IRssBuilder
    {
        Task<Rss> Query(Uri baseUrl, string feedId);
    }
}