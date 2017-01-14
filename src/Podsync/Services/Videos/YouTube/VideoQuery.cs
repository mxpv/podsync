using System.Collections.Generic;

namespace Podsync.Services.Videos.YouTube
{
    public struct VideoQuery
    {
        public ICollection<string> Ids { get; set; }
    }
}