using System;
using System.Linq;
using System.Security.Claims;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Podsync.Services;
using Podsync.Services.Resolver;

namespace Podsync.Helpers
{
    public static class Extensions
    {
        /// <summary>
        /// Generates a fully qualified URL to the specified content by using the specified content path.
        /// Converts a virtual (relative) path to an application absolute path.
        /// See http://stackoverflow.com/questions/30755827/getting-absolute-urls-using-asp-net-core-mvc-6
        /// </summary>
        /// <param name="url"></param>
        /// <param name="contentPath"></param>
        /// <returns></returns>
        public static string AbsoluteContent(this IUrlHelper url, string contentPath)
        {
            var request = url.ActionContext.HttpContext.Request;
            var baseUri = new Uri($"{request.Scheme}://{request.Host.Value}");
            var fullUri = new Uri(baseUri, url.Content(contentPath));

            return fullUri.ToString();
        }

        public static Uri GetBaseUrl(this HttpRequest request)
        {
            return new Uri($"{request.Scheme}://{request.Host}");
        }

        private const string OwnerId = "2822191";

        /// <summary>
        /// Check if user eligible for advanced features
        /// </summary>
        /// <param name="user"></param>
        /// <returns></returns>
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
            if (int.TryParse(user.GetClaim(Constants.Patreon.AmountDonated), out amount))
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

        public static string GetClaim(this ClaimsPrincipal claimsPrincipal, string type)
        {
            return claimsPrincipal.Claims.FirstOrDefault(x => x.Type == type)?.Value;
        }

        public static bool IsAudio(this ResolveFormat format)
        {
            return format == ResolveFormat.AudioHigh || format == ResolveFormat.AudioLow;
        }
    }
}