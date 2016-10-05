using System;

namespace Podsync.Services.Videos.YouTube
{
    public class YouTubeItem
    {
        public string VideoId { get; set; }

        public string ChannelId { get; set; }

        public string PlaylistId { get; set; }

        public Uri Link { get; set; }

        public Uri Thumbnail { get; set; }

        public string Title { get; set; }

        public string Description { get; set; }

        public DateTime PublishedAt { get; set; }
    }
}
