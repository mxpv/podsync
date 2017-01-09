using System;

namespace Podsync.Services.Videos.Vimeo
{
    public class Video
    {
        public string Id { get; set; }

        public string Title { get; set; }

        public string Description { get; set; }

        public Uri Link { get; set; }

        public Uri Thumbnail { get; set; }

        public DateTime CreatedAt { get; set; }

        public long Size { get; set; }

        public TimeSpan Duration { get; set; }

        public string Author { get; set; }
    }
}