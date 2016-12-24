using System.Linq;
using System.Security.Claims;
using Podsync.Services.Patreon;

namespace Podsync.Helpers
{
    public static class UserExtensions
    {
        private const string OwnerId = "2822191";

        public static bool EnablePatreonFeatures(this ClaimsPrincipal user)
        {
            if (!user.Identity.IsAuthenticated)
            {
                return false;
            }

            if (user.GetClaim(ClaimTypes.NameIdentifier) == OwnerId)
            {
                return true;
            }

            const int MinAmountCents = 100;

            int amount;
            if (int.TryParse(user.GetClaim(PatreonConstants.AmountDonated), out amount))
            {
                return amount >= MinAmountCents;
            }

            return false;
        }

        public static string GetName(this ClaimsPrincipal claimsPrincipal)
        {
            return claimsPrincipal.GetClaim(ClaimTypes.Name)
                   ?? claimsPrincipal.GetClaim(ClaimTypes.Email)
                   ?? "noname :(";
        }

        private static string GetClaim(this ClaimsPrincipal claimsPrincipal, string type)
        {
            return claimsPrincipal.Claims.FirstOrDefault(x => x.Type == type)?.Value;
        }
    }
}