using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.Threading.Tasks;
using System.Xml.Serialization;
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
        private const int DefaultPageSize = 50;

        private readonly XmlSerializer _serializer = new XmlSerializer(typeof(Rss));

        private readonly IRssBuilder _rssBuilder;
        private readonly ILinkService _linkService;
        private readonly IStorageService _storageService;

        public FeedController(IRssBuilder rssBuilder, ILinkService linkService, IStorageService storageService)
        {
            _rssBuilder = rssBuilder;
            _linkService = linkService;
            _storageService = storageService;
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
                Quality = request.Quality ?? ResolveType.VideoHigh,
                PageSize = request.PageSize ?? DefaultPageSize
            };

            if (!User.EnablePatreonFeatures())
            {
                feed.Quality = ResolveType.VideoHigh;
                feed.PageSize = DefaultPageSize;
            }

            var feedId = await _storageService.Save(feed);
            return _linkService.Feed(Request.GetBaseUrl(), feedId);
        }

        [HttpGet]
        [Route("~/{feedId:length(4, 6)}")]
        [ValidateModelState]
        public async Task<IActionResult> Feed([Required] string feedId)
        {
            Rss rss;

            try
            {
                rss = await _rssBuilder.Query(Request.GetBaseUrl(), feedId);
            }
            catch (KeyNotFoundException)
            {
                return NotFound(feedId);
            }

            // Set atom link to this feed
            // See https://validator.w3.org/feed/docs/warning/MissingAtomSelfLink.html
            var selfLink = new Uri($"{Request.Scheme}://{Request.Host}{Request.Path}");
            rss.Channels.ForEach(x => x.AtomLink = selfLink);

            // Serialize feed to string
            string body;
            using (var writer = new Utf8StringWriter())
            {
                _serializer.Serialize(writer, rss);
                body = writer.ToString();
            }

            return Content(body, "application/rss+xml; charset=UTF-8");
        }
    }
}