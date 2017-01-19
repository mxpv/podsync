using System;
using System.Threading.Tasks;
using Podsync.Services.Storage;

namespace Podsync.Services.Resolver
{
    public abstract class CachedResolver : IResolverService
    {
        private readonly TimeSpan UrlExpiration = TimeSpan.FromHours(3);
        private readonly IStorageService _storageService;

        protected CachedResolver(IStorageService storageService)
        {
            _storageService = storageService;
        }

        public abstract string Version { get; }

        public async Task<Uri> Resolve(Uri videoUrl, ResolveFormat format)
        {
            var id = videoUrl.GetHashCode().ToString();

            // Check if this video URL was resolved within last 3 hours
            var value = await _storageService.GetCached(Constants.Cache.VideosPrefix, id);
            if (!string.IsNullOrWhiteSpace(value))
            {
                return new Uri(value);
            }

            // Resolve and save to cache
            var uri = await ResolveInternal(videoUrl, format);
            await _storageService.Cache(Constants.Cache.VideosPrefix, id, uri.ToString(), UrlExpiration);

            return uri;
        }

        protected abstract Task<Uri> ResolveInternal(Uri videoUrl, ResolveFormat format);
    }
}