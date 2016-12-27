using Podsync.Services.Patreon;

namespace Podsync.Services
{
    public class PodsyncConfiguration
    {
        public string YouTubeApiKey { get; set; }

        public string RedisConnectionString { get; set; }

        public string PatreonClientId { get; set; }

        public string PatreonSecret { get; set; }

        public Tokens CreatorTokens { get; set; }
    }
}