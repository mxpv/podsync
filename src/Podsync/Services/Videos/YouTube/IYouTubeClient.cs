using System.Collections.Generic;
using System.Threading.Tasks;

namespace Podsync.Services.Videos.YouTube
{
    public interface IYouTubeClient
    {
        /// <summary>
        /// Returns a collection of zero or more channel resources that match the request criteria
        /// Cost: 5 (quota cost of 1 + id: 0, snippet: 2, contentDetails: 2)
        /// See https://developers.google.com/youtube/v3/docs/channels/list
        /// </summary>
        /// <param name="query"></param>
        /// <returns></returns>
        Task<IEnumerable<Channel>> GetChannels(ChannelQuery query);

        /// <summary>
        /// Returns a collection of playlists that match the API request parameters
        /// Cost: 3 (quota cost of 1 + id: 0, snippet: 2)
        /// See https://developers.google.com/youtube/v3/docs/playlists/list
        /// </summary>
        /// <param name="query"></param>
        /// <returns></returns>
        Task<IEnumerable<Playlist>> GetPlaylists(PlaylistQuery query);
        
        /// <summary>
        /// Returns a list of videos that match the API request parameters
        /// Cost: 5 (quota cost of 1 + id: 0, snippet: 2, contentDetails: 2)
        /// See https://developers.google.com/youtube/v3/docs/videos/list
        /// </summary>
        /// <param name="query"></param>
        /// <returns></returns>
        Task<IEnumerable<Video>> GetVideos(VideoQuery query);

        /// <summary>
        /// Returns a collection of playlist items that match the API request parameters.
        /// You can retrieve all of the playlist items in a specified playlist or 
        /// retrieve one or more playlist items by their unique IDs.
        /// Cost: 3 (quota cost of 1 + id: 0, snippet: 2)
        /// See https://developers.google.com/youtube/v3/docs/playlistItems/list
        /// </summary>
        /// <param name="query"></param>
        /// <returns></returns>
        Task<IEnumerable<Video>> GetPlaylistItems(PlaylistItemsQuery query);

        /// <summary>
        /// Optimized version of GetPlaylistItems to query video IDs only
        /// Cost: 3 (quota cost of 1 + id: 0, snippet: 2)
        /// </summary>
        /// <param name="query"></param>
        /// <returns></returns>
        Task<IEnumerable<string>> GetPlaylistItemIds(PlaylistItemsQuery query);
    }
}