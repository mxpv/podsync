using System.Net;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using Podsync.Services;
using Podsync.Services.Resolver;
using Xunit;

namespace Podsync.Tests.Controllers
{
    public class FeedControllerTests : TestServer<Startup>
    {
        [Fact]
        public async Task ValidateCreateNullTest()
        {
            var response = await Client.PostAsync("/feed/create", new StringContent("", Encoding.UTF8, "application/json"));
            Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        }

        [Theory]
        [InlineData(null, null, null)]
        [InlineData("a", null, null)]
        [InlineData("http://youtube.com", null, 0)]
        [InlineData("http://youtube.com", null, 25)]
        [InlineData("http://youtube.com", null, 151)]
        public async Task ValidateCreateTest(string url, ResolveFormat? quality, int? pageSize)
        {
            var feed = new CreateFeedRequest
            {
                Url = url,
                Quality = quality,
                PageSize = pageSize
            };

            var response = await Client.PostAsync("/feed/create", MakeHttpContent(feed));
            Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        }

        [Fact]
        public async Task CreateFeedTest()
        {
            var feed = new CreateFeedRequest
            {
                Url = "https://www.youtube.com/channel/UCKy1dAqELo0zrOtPkf0eTMw",
                Quality = ResolveFormat.AudioLow,
            };

            var response = await Client.PostAsync("/feed/create", MakeHttpContent(feed));
            Assert.Equal(HttpStatusCode.OK, response.StatusCode);

            var id = await response.Content.ReadAsStringAsync();
            Assert.NotNull(id);
        }
    }
}