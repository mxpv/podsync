using Podsync.Services.Resolver;

namespace Podsync.Services
{
    public static class Constants
    {
        public const int DefaultPageSize = 50;

        public const ResolveFormat DefaultFormat = ResolveFormat.VideoHigh;

        public static class Patreon
        {
            public const string AuthenticationScheme = "Patreon";

            public const string AuthorizationEndpoint = "https://www.patreon.com/oauth2/authorize";

            public const string TokenEndpoint = "https://api.patreon.com/oauth2/token";

            public const string AmountDonated = "Patreon/" + nameof(AmountDonated);
        }

        public static class Events
        {
            public const int RssError = 1;

            public const int YtdlError = 2;

            public const int UnhandledError = 3;
        }

        public static class Cache
        {
            public const string VideosPrefix = "video_urls";

            public const string FeedsPrefix = "feeds";
        }
    }
}