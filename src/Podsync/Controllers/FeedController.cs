using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.Linq;
using System.Security.Claims;
using System.Threading.Tasks;
using System.Xml.Serialization;
using Microsoft.ApplicationInsights;
using Microsoft.AspNetCore.Mvc;
using Podsync.Helpers;
using Podsync.Services;
using Podsync.Services.Builder;
using Podsync.Services.Feed;
using Podsync.Services.Feed.Internal;
using Podsync.Services.Links;
using Podsync.Services.Resolver;
using Podsync.Services.Storage;
using Shared;

namespace Podsync.Controllers
{
    [Route("feed")]
    [HandleException]
    public class FeedController : Controller
    {
        private static readonly IDictionary<string, string> Extensions = new Dictionary<string, string>
        {
            ["video/mp4"] = "mp4",
            ["audio/mp4"] = "m4a"
        };

        private readonly XmlSerializer _serializer = new XmlSerializer(typeof(Rss));

        private readonly IRssBuilder _rssBuilder;
        private readonly ILinkService _linkService;
        private readonly IStorageService _storageService;
        private readonly TelemetryClient _telemetry;

        public FeedController(IRssBuilder rssBuilder, ILinkService linkService, IStorageService storageService, TelemetryClient telemetry)
        {
            _rssBuilder = rssBuilder;
            _linkService = linkService;
            _storageService = storageService;
            _telemetry = telemetry;
        }

        [HttpPost]
        [Route("create")]
        [ValidateModelState]
        public async Task<Uri> Create([FromBody] CreateFeedRequest request)
        {
            var linkInfo = _linkService.Parse(new Uri(request.Url));

            if (linkInfo.LinkType == LinkType.Video)
            {
                throw new ArgumentException("Direct links are not supported, you should provide group, channel or user link");
            }

            var feed = new FeedMetadata
            {
                Provider = linkInfo.Provider,
                LinkType = linkInfo.LinkType,
                Id = linkInfo.Id,
                Quality = request.Quality ?? ResolveType.VideoHigh,
                PageSize = request.PageSize ?? Constants.DefaultPageSize
            };

            // Check if user eligible for Patreon features
            var enablePatreonFeatures = User.EnablePatreonFeatures();
            if (!enablePatreonFeatures)
            {
                feed.Quality = ResolveType.VideoHigh;
                feed.PageSize = Constants.DefaultPageSize;
            }

            var feedId = await _storageService.Save(feed);
            var url = _linkService.Feed(Request.GetBaseUrl(), feedId);

            // Report metrics
            var properties = new Dictionary<string, string>
            {
                ["Provider"] = linkInfo.Provider.ToString(),
                ["Patreon"] = enablePatreonFeatures.ToString(),
                ["Format"] = feed.Quality == ResolveType.AudioHigh || feed.Quality == ResolveType.AudioLow ? "Audio" : "Video",
                ["Quality"] = feed.Quality == ResolveType.AudioHigh || feed.Quality == ResolveType.VideoHigh ? "Hight" : "Low",
                ["PageSize"] = feed.PageSize.ToString()
            };

            if (User.Identity.IsAuthenticated)
            {
                properties.Add("User", User.GetClaim(ClaimTypes.NameIdentifier));
                properties.Add("Email", User.GetClaim(ClaimTypes.Email));
            }

            _telemetry.TrackEvent("CreateFeed", properties);

            return url;
        }

        [HttpGet]
        [Route("~/{feedId:length(4, 6)}")]
        [ValidateModelState]
        public async Task<IActionResult> Feed([Required] string feedId)
        {
            Rss rss;

            try
            {
                rss = await _rssBuilder.Query(feedId);
            }
            catch (KeyNotFoundException)
            {
                return NotFound(feedId);
            }

            var selfHost = Request.GetBaseUrl();

            // Set atom link to this feed
            // See https://validator.w3.org/feed/docs/warning/MissingAtomSelfLink.html
            var selfLink = new Uri(selfHost, Request.Path);
            rss.Channels.ForEach(x => x.AtomLink = selfLink);

            // No magic here, just make download links to DownloadController.Download
            rss.Channels.SelectMany(x => x.Items).ForEach(item =>
            {
                var ext = Extensions[item.ContentType];
                item.DownloadLink = new Uri(selfHost, $"download/{feedId}/{item.Id}.{ext}");
            });

            // Serialize feed to string
            string body;
            using (var writer = new Utf8StringWriter())
            {
                _serializer.Serialize(writer, rss);
                body = writer.ToString();
            }

            // Report metrics
            _telemetry.TrackEvent("GetFeed");

            return Content(body, "application/rss+xml; charset=UTF-8");
        }
    }
}