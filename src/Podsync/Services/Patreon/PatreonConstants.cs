namespace Podsync.Services.Patreon
{
    public static class PatreonConstants
    {
        public const string AuthenticationScheme = "Patreon";

        public const string AuthorizationEndpoint = "https://www.patreon.com/oauth2/authorize";

        public const string TokenEndpoint = "https://api.patreon.com/oauth2/token";

        public const string AmountDonated = "Patreon/" + nameof(AmountDonated);
    }
}