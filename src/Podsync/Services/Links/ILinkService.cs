using System;

namespace Podsync.Services.Links
{
    public interface ILinkService
    {
        LinkInfo Parse(Uri link);

        Uri Make(LinkInfo info);

        Uri Feed(Uri baseUrl, string feedId);
    }
}