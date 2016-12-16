using System.Threading.Tasks;
using Xunit;

namespace Podsync.Tests.Controllers
{
    public class StatusControllerTests : TestServer<Startup>
    {
        [Fact]
        public async Task StatusTest()
        {
            var response = await Client.GetAsync("/status");
            var content = await response.Content.ReadAsStringAsync();

            Assert.Contains("Path: /status", content);
        }
    }
}