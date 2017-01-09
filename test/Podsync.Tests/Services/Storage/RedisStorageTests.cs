using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Podsync.Services.Links;
using Podsync.Services.Storage;
using Xunit;

namespace Podsync.Tests.Services.Storage
{
    public class RedisStorageTests : TestBase, IDisposable
    {
        private readonly RedisStorage _storage;

        public RedisStorageTests()
        {
            _storage = new RedisStorage(Options);
        }

        [Fact]
        public async Task PingTest()
        {
            var time = await _storage.Ping();
            Assert.True(time.TotalMilliseconds > 0);
        }

        [Fact]
        public void MakeIdTest()
        {
            const int idCount = 50;


            var results = new string[idCount];
            Parallel.For(0, results.Length, (i, _) => results[i] = _storage.MakeId().GetAwaiter().GetResult());

            Assert.Equal(results.Length, results.Distinct().Count());
        }

        [Fact]
        public async Task SaveLoadFeedTest()
        {
            var feed = new FeedMetadata
            {
                Id = "123",
                LinkType = LinkType.Channel,
                Provider = Provider.Vimeo,

                PageSize = 45
            };

            var id = await _storage.Save(feed);
            Assert.NotEmpty(id);
            Assert.Equal(4, id.Length);

            var loaded = await _storage.Load(id);

            Assert.Equal(feed.Id, loaded.Id);
            Assert.Equal(feed.LinkType, loaded.LinkType);
            Assert.Equal(feed.Provider, loaded.Provider);
            Assert.Equal(45, loaded.PageSize);
        }

        [Fact]
        public Task LoadInvalidFeedTest()
        {
            return Assert.ThrowsAsync<KeyNotFoundException>(() => _storage.Load("test"));
        }

        public void Dispose()
        {
            _storage.Dispose();
        }
    }
}