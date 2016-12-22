using System;
using Microsoft.AspNetCore.Mvc;

namespace Podsync.Helpers
{
    public static class UrlExtensions
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
    }
}