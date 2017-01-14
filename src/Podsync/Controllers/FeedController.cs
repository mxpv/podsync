using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.Linq;
using System.Security.Claims;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Podsync.Helpers;
using Podsync.Services;
using Podsync.Services.Links;
using Podsync.Services.Rss;
using Podsync.Services.Rss.Contracts;
using Podsync.Services.Storage;
using Shared;

namespace Podsync.Controllers
{
    [Route("feed")]
    [ServiceFilter(typeof(HandleExceptionAttribute), IsReusable = true)]
    public class FeedController : Controller
    {
        private static readonly IDictionary<string, string> Extensions = new Dictionary<string, string>
        {
            ["video/mp4"] = "mp4",
            ["audio/mp4"] = "m4a"
        };

        private readonly ILinkService _linkService;
        private readonly IFeedService _feedService;

        public FeedController(IRssBuilder rssBuilder, ILinkService linkService, IStorageService storageService, IFeedService feedService)
        {
            _linkService = linkService;
            _feedService = feedService;
        }

        [HttpPost]
        [Route("create")]
        [ValidateModelState]
        public async Task<Uri> Create([FromBody] CreateFeedRequest request)
        {
            var linkInfo = _linkService.Parse(new Uri(request.Url));

            var feed = new FeedMetadata
            {
                Provider = linkInfo.Provider,
                LinkType = linkInfo.LinkType,
                Id = linkInfo.Id,
                Quality = request.Quality ?? Constants.DefaultFormat,
                PageSize = request.PageSize ?? Constants.DefaultPageSize
            };

            // Check if user eligible for Patreon features
            var allowFeatures = User.EnablePatreonFeatures();
            if (allowFeatures)
            {
                feed.PatreonId = User.GetClaim(ClaimTypes.NameIdentifier);
            }
            else
            {
                feed.Quality = Constants.DefaultFormat;
                feed.PageSize = Constants.DefaultPageSize;
            }

            var feedId = await _feedService.Create(feed);
            var url = _linkService.Feed(Request.GetBaseUrl(), feedId);

            return url;
        }

        [HttpGet]
        [Route("~/{feedId:length(4, 6)}")]
        [ValidateModelState]
        public async Task<IActionResult> Feed([Required] string feedId)
        {
            Feed feed;

            try
            {
                feed = await _feedService.Get(feedId);
            }
            catch (KeyNotFoundException)
            {
                return NotFound($"ERROR: No feed with id {feedId}");
            }

            var selfHost = Request.GetBaseUrl();

            // Set atom link to this feed
            // See https://validator.w3.org/feed/docs/warning/MissingAtomSelfLink.html
            var selfLink = new Uri(selfHost, Request.Path);
            feed.Channels.ForEach(x => x.AtomLink = selfLink);

            // No magic here, just make download links to DownloadController.Download
            feed.Channels.SelectMany(x => x.Items).ForEach(item =>
            {
                var ext = Extensions[item.ContentType];
                item.DownloadLink = new Uri(selfHost, $"download/{feedId}/{item.Id}.{ext}");
            });

            return Content(feed.ToString(), "application/rss+xml; charset=UTF-8");
        }
    }
}