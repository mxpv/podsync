using System;
using System.Threading.Tasks;

namespace Podsync.Services.Patreon
{
    public interface IPatreonApi : IDisposable
    {
        Task<dynamic> FetchProfile(Tokens tokens);
    }
}