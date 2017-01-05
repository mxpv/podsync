using System.Collections.Generic;
using System.Threading.Tasks;

namespace Podsync.Services.Videos.Vimeo
{
    public interface IVimeoClient
    {
        Task<Group> Group(string id);

        Task<Group> Channel(string id);

        Task<User> User(string id);

        Task<IEnumerable<Video>> GroupVideos(string id, int count);

        Task<IEnumerable<Video>> UserVideos(string id, int count);

        Task<IEnumerable<Video>> ChannelVideos(string id, int count);
    }
}