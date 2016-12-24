using System.Threading.Tasks;
using Podsync.Services.Patreon;
using Xunit;

namespace Podsync.Tests.Services.Patreon
{
    public class PatreonApiTests : TestBase
    {
        private readonly IPatreonApi _api = new PatreonApi();

        private Tokens Tokens => Configuration.CreatorTokens;

        [Fact]
        public async Task FetchProfileTest()
        {
            var user = await _api.FetchUserAndPledges(Tokens);

            Assert.Equal("2822191", user.Id);
            Assert.Equal("pavlenko.maksym@gmail.com", user.Email);
            Assert.Equal("https://www.patreon.com/podsync", user.Url);
            Assert.Equal("Max", user.Name);
        }
    }
}