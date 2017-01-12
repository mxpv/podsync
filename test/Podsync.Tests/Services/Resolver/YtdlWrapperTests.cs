using System;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Moq;
using Podsync.Services.Resolver;
using Podsync.Services.Storage;
using Xunit;

namespace Podsync.Tests.Services.Resolver
{
    public class YtdlWrapperTests : TestBase
    {
        private readonly Mock<ILogger<YtdlWrapper>> _logger = new Mock<ILogger<YtdlWrapper>>();
        private readonly Mock<IStorageService> _storage = new Mock<IStorageService>();

        private readonly IResolverService _resolver;

        public YtdlWrapperTests()
        {
            _storage.Setup(x => x.GetCached(It.IsAny<string>(), It.IsAny<string>())).ReturnsAsync("");
            _resolver = new YtdlWrapper(_storage.Object, _logger.Object);
        }

        [Theory]
        [InlineData("https://www.youtube.com/watch?v=BaW_jenozKc")]
        public async Task ResolveTest(string url)
        {
            _storage.ResetCalls();

            var videoUrl = new Uri(url);
            var downloadUrl = await _resolver.Resolve(videoUrl);

            _storage.Verify(x => x.GetCached("video_urls", videoUrl.GetHashCode().ToString()), Times.Once);
            _storage.Verify(x => x.Cache("video_urls", videoUrl.GetHashCode().ToString(), It.IsAny<string>(), It.IsAny<TimeSpan>()), Times.Once);

            Assert.NotEqual(downloadUrl, videoUrl);
            Assert.True(downloadUrl.IsAbsoluteUri);
        }

        [Theory]
        [InlineData("https://www.youtube.com/watch?v=fiWMUkOgY9I")]
        public async Task FailTest(string url)
        {
            var ex = await Assert.ThrowsAsync<InvalidOperationException>(async () => await _resolver.Resolve(new Uri(url)));
            Assert.NotEmpty(ex.Message);
        }

        [Fact]
        public void VersionTest()
        {
            Assert.NotNull(_resolver.Version);
        }

        [Fact]
        public async Task ResolveOutputTest()
        {
            var downloadUrl = await _resolver.Resolve(new Uri("https://www.youtube.com/watch?v=-csRxRj_zcw&t=45s"), ResolveFormat.AudioHigh);
            Assert.True(downloadUrl.IsAbsoluteUri);
        }
    }
}