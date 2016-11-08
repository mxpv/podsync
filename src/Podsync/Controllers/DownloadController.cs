using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Podsync.Services.Links;
using Podsync.Services.Resolver;

namespace Podsync.Controllers
{
    [Route("download")]
    public class DownloadController : Controller
    {
        private readonly IResolverService _resolverService;
        private readonly ILinkService _linkService;

        public DownloadController(IResolverService resolverService, ILinkService linkService)
        {
            _resolverService = resolverService;
            _linkService = linkService;
        }

        [Route("{provider}/{videoId}.mp4")]
        public async Task<IActionResult> Download(Provider provider, string videoId)
        {
            var url = _linkService.Make(new LinkInfo
            {
                Provider = provider,
                LinkType = LinkType.Video,
                Id = videoId
            });

            var redirectUrl = await _resolverService.Resolve(url);

            return Redirect(redirectUrl.ToString());
        }
    }
}