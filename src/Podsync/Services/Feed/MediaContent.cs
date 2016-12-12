using System;

namespace Podsync.Services.Feed
{
    public struct MediaContent
    {
        public Uri Url { get; set; }

        public long Length { get; set; }

        public string MediaType { get; set; }
    }
}