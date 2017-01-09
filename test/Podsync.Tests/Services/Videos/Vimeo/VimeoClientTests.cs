using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Podsync.Services.Videos.Vimeo;
using Xunit;

// ReSharper disable PossibleMultipleEnumeration
namespace Podsync.Tests.Services.Videos.Vimeo
{
    public class VimeoClientTests : TestBase, IDisposable
    {
        private readonly VimeoClient _client;

        public VimeoClientTests()
        {
            _client = new VimeoClient(Options);
        }
        
        [Fact]
        public async Task ChannelTest()
        {
            var channel = await _client.Channel("staffpicks");

            Assert.Equal(new Uri("https://vimeo.com/channels/staffpicks"), channel.Link);
            Assert.Equal("Vimeo Staff Picks", channel.Name);
            Assert.Equal("Vimeo Curation", channel.Author);
            Assert.False(string.IsNullOrWhiteSpace(channel.Description));
        }

        [Fact]
        public async Task GroupTest()
        {
            var group = await _client.Group("motion");

            Assert.Equal(new Uri("https://vimeo.com/groups/motion"), group.Link);
            Assert.Equal("Motion Graphic Artists", group.Name);
            Assert.Equal("Danny Garcia", group.Author);
            Assert.False(string.IsNullOrWhiteSpace(group.Description));
        }

        [Fact]
        public async Task UserTest()
        {
            var user = await _client.User("motionarray");

            Assert.Equal("Motion Array", user.Name);
            Assert.False(string.IsNullOrWhiteSpace(user.Bio));
            Assert.Equal(new Uri("https://vimeo.com/motionarray"), user.Link);
        }

        [Fact]
        public async Task GroupVideosTest()
        {
            var videos = await _client.GroupVideos("motion", 101);
            Assert.Equal(101, videos.Count());
            ValidateCollection(videos);
        }

        [Fact]
        public async Task UserVideosTest()
        {
            var videos = await _client.UserVideos("motionarray", 7);
            Assert.Equal(7, videos.Count());
            ValidateCollection(videos);
        }

        [Fact]
        public async Task ChannelVideosTest()
        {
            var videos = await _client.ChannelVideos("staffpicks", 44);
            Assert.Equal(44, videos.Count());
            ValidateCollection(videos);
        }

        public void Dispose()
        {
            _client.Dispose();
        }

        private void ValidateCollection(IEnumerable<Video> videos)
        {
            foreach (var video in videos)
            {
                Assert.False(string.IsNullOrWhiteSpace(video.Title));
                Assert.True(video.Duration.TotalSeconds > 1);
                Assert.True(video.Size > 0);
                Assert.NotNull(video.Link);
            }
        }
    }
}