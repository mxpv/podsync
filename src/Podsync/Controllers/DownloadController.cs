using System.Threading.Tasks;
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

        public DownloadController(IResolverService resolverService, ILinkService linkService, IStorageService storageService)
        {
            _resolverService = resolverService;
            _linkService = linkService;
            _storageService = storageService;
        }
        
        // Main video download endpoint, don't forget to reflect any changes in LinkService.Download
        [HttpGet]
        [Route("{feedId}/{videoId}/")]
        public async Task<IActionResult> Download(string feedId, string videoId)
        {
            var metadata = await _storageService.Load(feedId);

            var url = _linkService.Make(new LinkInfo
            {
                Provider = metadata.Provider,
                LinkType = LinkType.Video,
                Id = videoId
            });

            var redirectUrl = await _resolverService.Resolve(url, metadata.Quality);

            return Redirect(redirectUrl.ToString());
        }
    }
}