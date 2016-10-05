using System;

namespace Podsync.Services.Videos.YouTube
{
    public class Video : YouTubeItem
    {
        public TimeSpan Duration { get; set; }

        public long Size { get; set; }
    }
}