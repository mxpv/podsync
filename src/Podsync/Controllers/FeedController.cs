using System.IO;
using System.Threading.Tasks;
using System.Xml.Serialization;
using Microsoft.AspNetCore.Mvc;
using Podsync.Services.Builder;
using Podsync.Services.Feed;

namespace Podsync.Controllers
{
    [Route("feed")]
    public class FeedController : Controller
    {
        private readonly XmlSerializer _serializer = new XmlSerializer(typeof(Rss));

        private readonly IRssBuilder _rssBuilder;

        public FeedController(IRssBuilder rssBuilder)
        {
            _rssBuilder = rssBuilder;
        }

        [Route("{feedId}")]
        public async  Task<IActionResult> Index(string feedId)
        {
            var rss = await _rssBuilder.Query(feedId);

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