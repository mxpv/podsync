using Podsync.Services.Resolver;

namespace Podsync.Services
{
    public static class Constants
    {
        public const int DefaultPageSize = 50;

        public const ResolveType DefaultFormat = ResolveType.VideoHigh;

        public static class Patreon
        {
            public const string AuthenticationScheme = "Patreon";

            public const string AuthorizationEndpoint = "https://www.patreon.com/oauth2/authorize";

            public const string TokenEndpoint = "https://api.patreon.com/oauth2/token";

            public const string AmountDonated = "Patreon/" + nameof(AmountDonated);
        }
    }
}