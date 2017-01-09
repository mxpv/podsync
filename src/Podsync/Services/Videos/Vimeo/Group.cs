using System;

namespace Podsync.Services.Videos.Vimeo
{
    public class Group
    {
        public string Name { get; set; }

        public string Description { get; set; }

        public Uri Link { get; set; }

        public DateTime CreatedAt { get; set; }

        public Uri Thumbnail { get; set; }

        public string Author { get; set; }
    }
}