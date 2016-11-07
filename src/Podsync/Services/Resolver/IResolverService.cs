using System;
using System.Threading.Tasks;

namespace Podsync.Services.Resolver
{
    public interface IResolverService
    {
        Task<Uri> Resolve(Uri videoUrl, FileType fileType = FileType.Video, Quality quality = Quality.High);
    }
}