using System;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Moq;
using Podsync.Services.Resolver;
using Xunit;

namespace Podsync.Tests.Services.Resolver
{
    public class YtdlWrapperTests : TestBase
    {
        private readonly IMock<ILogger<YtdlWrapper>> _logger = new Mock<ILogger<YtdlWrapper>>();
        private readonly IResolverService _resolver;

        public YtdlWrapperTests()
        {
            _resolver = new YtdlWrapper(_logger.Object);
        }

        [Theory]
        [InlineData("https://www.youtube.com/watch?v=BaW_jenozKc")]
        public async Task ResolveTest(string url)
        {
            var videoUrl = new Uri(url);
            var downloadUrl = await _resolver.Resolve(videoUrl);

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
            var downloadUrl = await _resolver.Resolve(new Uri("https://www.youtube.com/watch?v=-csRxRj_zcw&t=45s"), ResolveType.AudioHigh);
            Assert.True(downloadUrl.IsAbsoluteUri);
        }
    }
}