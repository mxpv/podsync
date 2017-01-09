using System;
using System.Threading.Tasks;

namespace Podsync.Services.Resolver
{
    public interface IResolverService
    {
        string Version { get; }

        Task<Uri> Resolve(Uri videoUrl, ResolveFormat format = ResolveFormat.VideoHigh);
    }
}