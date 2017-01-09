using System;

namespace Podsync.Services.Videos.Vimeo
{
    public class User
    {
        public string Name { get; set; }

        public string Bio { get; set; }

        public DateTime CreatedAt { get; set; }

        public Uri Link { get; set; }

        public Uri Thumbnail { get; set; }
    }
}