namespace Podsync.Services.Videos.YouTube
{
    public struct ChannelQuery
    {
        public string ChannelId { get; set; }

        public string Username { get; set; }

        public int? Count { get; set; }
    }
}