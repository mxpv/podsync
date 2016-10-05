namespace Podsync.Services.Videos.YouTube
{
    public struct PlaylistItemsQuery
    {
        /// <summary>
        /// The id parameter specifies a comma-separated list of one or more unique playlist item IDs
        /// </summary>
        public string Id { get; set; }

        /// <summary>
        /// The playlistId parameter specifies the unique ID of the playlist for which you want to retrieve playlist items.
        /// Note that even though this is an optional parameter, every request to retrieve playlist items must specify 
        /// a value for either the id parameter or the playlistId parameter.
        /// </summary>
        public string PlaylistId { get; set; }

        /// <summary>
        /// The videoId parameter specifies that the request should return only the playlist items that contain the specified video
        /// </summary>
        public string VideoId { get; set; }

        public uint? Count { get; set; }
    }
}