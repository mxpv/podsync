using System;

namespace Podsync.Services.Links
{
    public interface ILinkService
    {
        LinkInfo Parse(Uri link);

        Uri Make(LinkInfo info);

        Uri Download(string feedId, string videoId);
    }
}