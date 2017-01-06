using System;
using System.Linq;
using System.Threading.Tasks;
using Moq;
using Podsync.Services.Builder;
using Podsync.Services.Links;
using Podsync.Services.Storage;
using Podsync.Services.Videos.Vimeo;
using Xunit;

namespace Podsync.Tests.Services.Builder
{
    public class VimeoRssBuilderTests : TestBase
    {
        private readonly Mock<IStorageService> _storageService = new Mock<IStorageService>();

        private readonly VimeoRssBuilder _builder;

        public VimeoRssBuilderTests()
        {
            _builder = new VimeoRssBuilder(_storageService.Object, new VimeoClient(Options));
        }

        [Theory]
        [InlineData(LinkType.Channel, "staffpicks")]
        [InlineData(LinkType.Group, "motion")]
        [InlineData(LinkType.User, "motionarray")]
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

            var rss = await _builder.Query(new Uri("http://localhost:2020"), feedId);

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
                Assert.NotNull(item.Content);
                Assert.True(item.Content.Length > 0);
                Assert.NotNull(item.Content.MediaType);
                Assert.NotNull(item.PubDate);
            }
        }
    }
}