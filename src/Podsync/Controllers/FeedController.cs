using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations;
using System.IO;
using System.Threading.Tasks;
using System.Xml.Serialization;
using Microsoft.AspNetCore.Mvc;
using Podsync.Helpers;
using Podsync.Services;
using Podsync.Services.Builder;
using Podsync.Services.Feed;
using Podsync.Services.Links;
using Podsync.Services.Resolver;
using Podsync.Services.Storage;

namespace Podsync.Controllers
{
    [Route("feed")]
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
        public Task<string> Create([FromBody] CreateFeedRequest request)
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

            return _storageService.Save(feed);
        }

        [HttpGet]
        [Route("{feedId}")]
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

            // Serialize feed to string
            string body;
            using (var writer = new StringWriter())
            {
                _serializer.Serialize(writer, rss);
                body = writer.ToString();
            }

            return Content(body, "text/xml");
        }
    }
}