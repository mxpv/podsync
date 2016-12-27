using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using System.Linq;
using Podsync.Services.Links;
using Podsync.Services.Videos.YouTube;
using Xunit;

namespace Podsync.Tests.Services.Videos.YouTube
{
    public class YouTubeClientTests : TestBase, IDisposable
    {
        private readonly YouTubeClient _client;

        public YouTubeClientTests()
        {
            _client = new YouTubeClient(new LinkService(), Options);
        }

        [Fact]
        public async Task GetChannelTest()
        {
            var query = new ChannelQuery
            {
                ChannelId = "UC0JB7TSe49lg56u6qH8y_MQ"
            };

            var list = await _client.GetChannels(query);
            var channel = list.Single();

            Assert.Equal(query.ChannelId, channel.ChannelId);
            Assert.Equal("GDC", channel.Title);
            Assert.False(string.IsNullOrEmpty(channel.Description));
            Assert.NotEqual(DateTime.MinValue, channel.PublishedAt);
            Assert.NotNull(channel.Thumbnail);
        }

        [Fact]
        public async Task GetPlaylistsTest()
        {
            var query = new PlaylistQuery
            {
                PlaylistId = "PL2e4mYbwSTbbiX2uwspn0xiYb8_P_cTAr",
                Count = 1
            };

            var list = await _client.GetPlaylists(query);
            var playlist = list.Single();
            
            Assert.Equal("PL2e4mYbwSTbbiX2uwspn0xiYb8_P_cTAr", playlist.PlaylistId);
            Assert.Equal("UC0JB7TSe49lg56u6qH8y_MQ", playlist.ChannelId);
            Assert.Equal("GDC: Postmortems", playlist.Title);
            Assert.Equal(new Uri("https://youtube.com/playlist?list=PL2e4mYbwSTbbiX2uwspn0xiYb8_P_cTAr"), playlist.Link);
            Assert.NotNull(playlist.Thumbnail);

            Assert.Equal(new DateTime(2015, 06, 29), playlist.PublishedAt.Date);
        }

        [Fact]
        public async Task GetPlaylistIdsTest()
        {
            var query = new PlaylistItemsQuery
            {
                PlaylistId = "PL2e4mYbwSTbbiX2uwspn0xiYb8_P_cTAr",
                Count = 3
            };

            var list = await _client.GetPlaylistItemIds(query);
            Assert.NotEmpty(list);
        }

        [Fact]
        public async Task GetVideosTest()
        {
            var query = new VideoQuery
            {
                Id = "OlYH4gDi0Sk,kkcKnWrCZ7k"
            };

            var response = await _client.GetVideos(query);
            var list = response as IList<Video> ?? response.ToList();

            Assert.Equal(2, list.Count);

            var last = list.Last();

            Assert.Equal("kkcKnWrCZ7k", last.VideoId);
            Assert.Equal("UCS-aAvZsVegeMDfpdCPjWnA", last.ChannelId);
            Assert.Equal("Deus Ex: Mankind Divided - Full Soundtrack OST", last.Title);
            Assert.False(string.IsNullOrEmpty(last.Description));
            Assert.Equal(new TimeSpan(0, 1, 57, 28), last.Duration);
            Assert.Equal(new DateTime(2016, 8, 24), last.PublishedAt.Date);
            Assert.Equal(new Uri("https://youtube.com/watch?v=kkcKnWrCZ7k"), last.Link);
            Assert.True(last.Size > 0);
        }

        [Fact]
        public async Task GetPlaylistItemsTest()
        {
            var query = new PlaylistItemsQuery
            {
                PlaylistId = "PLWG9qzvMcoJGw9AS0XgLfrwDR2kcAMLhs",
                Count = 10
            };

            var response = await _client.GetPlaylistItems(query);
            var list = response as IList<Video> ?? response.ToList();

            Assert.Equal(4, list.Count);

            foreach (var video in list)
            {
                Assert.False(string.IsNullOrEmpty(video.VideoId));
                Assert.False(string.IsNullOrEmpty(video.ChannelId));
                Assert.False(string.IsNullOrEmpty(video.PlaylistId));

                Assert.False(string.IsNullOrEmpty(video.Title));
            }
        }

        public void Dispose()
        {
            _client.Dispose();
        }
    }
}