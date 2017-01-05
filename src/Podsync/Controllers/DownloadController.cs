using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Microsoft.ApplicationInsights;
using Microsoft.AspNetCore.Mvc;
using Podsync.Helpers;
using Podsync.Services.Links;
using Podsync.Services.Resolver;
using Podsync.Services.Storage;

namespace Podsync.Controllers
{
    [Route("download")]
    [HandleException]
    public class DownloadController : Controller
    {
        private readonly IResolverService _resolverService;
        private readonly ILinkService _linkService;
        private readonly IStorageService _storageService;
        private readonly TelemetryClient _telemetry;

        public DownloadController(IResolverService resolverService, ILinkService linkService, IStorageService storageService, TelemetryClient telemetry)
        {
            _resolverService = resolverService;
            _linkService = linkService;
            _storageService = storageService;
            _telemetry = telemetry;
        }
        
        // Main video download endpoint, don't forget to reflect any changes in LinkService.Download
        [HttpGet]
        [Route("{feedId}/{videoId}.{ext:length(3,3)}")]
        public async Task<IActionResult> Download(string feedId, string videoId)
        {
            var metadata = await _storageService.Load(feedId);

            var url = _linkService.Make(new LinkInfo
            {
                Provider = metadata.Provider,
                LinkType = LinkType.Video,
                Id = videoId
            });

            Uri redirectUrl;

            try
            {
                redirectUrl = await _resolverService.Resolve(url, metadata.Quality);
            }
            catch (Exception ex)
            {
                _telemetry.TrackException(ex, new Dictionary<string, string>
                {
                    ["FeedId"] = feedId,
                    ["VideoId"] = videoId
                });

                var response = "Could nou resolve URL";
                if (ex is InvalidOperationException)
                {
                    response = ex.Message;
                }

                return BadRequest(response);
            }

            // Report metrics
            _telemetry.TrackEvent("Download");

            return Redirect(redirectUrl.ToString());
        }
    }
}