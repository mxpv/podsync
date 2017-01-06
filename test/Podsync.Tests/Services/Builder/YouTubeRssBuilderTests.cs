using System;
using System.Linq;
using System.Threading.Tasks;
using Moq;
using Podsync.Services.Builder;
using Podsync.Services.Links;
using Podsync.Services.Storage;
using Podsync.Services.Videos.YouTube;
using Xunit;

namespace Podsync.Tests.Services.Builder
{
    public class YouTubeRssBuilderTests : TestBase
    {
        private readonly Mock<IStorageService> _storageService = new Mock<IStorageService>();

        private readonly YouTubeRssBuilder _builder;

        public YouTubeRssBuilderTests()
        {
            var linkService = new LinkService();
            var client = new YouTubeClient(linkService, Options);

            _builder = new YouTubeRssBuilder(client, _storageService.Object);
        }

        [Theory]
        [InlineData(LinkType.Channel, "UC0JB7TSe49lg56u6qH8y_MQ")]
        [InlineData(LinkType.User, "fxigr1")]
        [InlineData(LinkType.Playlist, "PL2e4mYbwSTbbiX2uwspn0xiYb8_P_cTAr")]
        public async Task BuildRssTest(LinkType linkType, string id)
        {
            var feed = new FeedMetadata
            {
                Provider = Provider.YouTube,
                LinkType = linkType,
                Id = id
            };

            var feedId = DateTime.UtcNow.Ticks.ToString();

            _storageService.Setup(x => x.Load(feedId)).ReturnsAsync(feed);

            var rss = await _builder.Query(feedId);

            Assert.NotEmpty(rss.Channels);

            var channel = rss.Channels.Single();

            Assert.NotNull(channel.Title);
            Assert.NotNull(channel.Description);
            Assert.NotNull(channel.Image);
            Assert.NotNull(channel.Guid);
            Assert.NotEmpty(channel.Items);

            foreach (var item in channel.Items)
            {
                Assert.NotNull(item.Title);
                Assert.NotNull(item.Link);
                Assert.True(item.Duration.TotalSeconds > 0);
                Assert.True(item.FileSize > 0);
                Assert.NotNull(item.ContentType);
                Assert.NotNull(item.PubDate);
            }
        }
    }
}