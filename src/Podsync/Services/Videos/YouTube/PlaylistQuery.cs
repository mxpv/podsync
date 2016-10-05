namespace Podsync.Services.Videos.YouTube
{
    public struct PlaylistQuery
    {
        public string PlaylistId { get; set; }

        public string ChannelId { get; set; }

        public uint? Count { get; set; }
    }
}