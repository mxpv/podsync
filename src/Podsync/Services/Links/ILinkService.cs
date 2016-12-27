using System;

namespace Podsync.Services.Links
{
    public interface ILinkService
    {
        LinkInfo Parse(Uri link);

        Uri Make(LinkInfo info);

        Uri Download(Uri baseUrl, string feedId, string videoId, string ext);

        Uri Feed(Uri baseUrl, string feedId);
    }
}