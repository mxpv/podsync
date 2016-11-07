using System;
using System.Threading.Tasks;
using Podsync.Services.Resolver;
using Xunit;

namespace Podsync.Tests.Services.Resolver
{
    public class YtdlWrapperTests
    {
        private readonly IResolverService _resolver = new YtdlWrapper();

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
    }
}