using System;
using System.Net.Http;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using Podsync.Services.Storage;

namespace Podsync.Services.Resolver
{
    public class RemoteResolver : CachedResolver
    {
        private readonly ILogger _logger;
        private readonly HttpClient _client = new HttpClient();

        public RemoteResolver(IStorageService storageService, IOptions<PodsyncConfiguration> options, ILogger<RemoteResolver> logger) : base(storageService)
        {
            _logger = logger;
            _client.BaseAddress = new Uri(options.Value.RemoteResolverUrl);

            _logger.LogInformation($"Remote resolver URL: {_client.BaseAddress}");
        }

        public override string Version { get; } = "Remote";

        protected override async Task<Uri> ResolveInternal(Uri videoUrl, ResolveFormat format)
        {
            using(var response = await _client.GetAsync($"/resolve?url={videoUrl}&quality={format}"))
            {
                using(response.Content)
                {
                    var body = await response.Content.ReadAsStringAsync();
                    if (!response.IsSuccessStatusCode)
                    {
                        throw new HttpRequestException(body);
                    }

                    return new Uri(body);
                }
            }
        }
    }
}