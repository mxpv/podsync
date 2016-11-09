using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using HashidsNet;
using Microsoft.Extensions.Options;
using Podsync.Services.Links;
using StackExchange.Redis;

namespace Podsync.Services.Storage
{
    public class RedisStorage : IStorageService
    {
        private const string IdKey = "keygen";
        private const string IdSalt = "65fce519433f4218aa0cee6394225eea";
        private const int IdLength = 4;

        // Store all fields manually for backward compatibility with existing implementation
        private const string ProviderField = "provider";
        private const string TypeField = "type";
        private const string IdField = "id";

        private static readonly IHashids HashIds = new Hashids(IdSalt, IdLength);

        private readonly IDatabase _db;

        public RedisStorage(IOptions<PodsyncConfiguration> configuration)
        {
            var cs = configuration.Value.RedisConnectionString;
            var connection = ConnectionMultiplexer.ConnectAsync(cs).GetAwaiter().GetResult();

            _db = connection.GetDatabase();
        }

        public void Dispose()
        {
            _db.Multiplexer.Dispose();
        }

        public Task<TimeSpan> Ping()
        {
            return _db.PingAsync();
        }

        public async Task<string> Save(FeedMetadata metadata)
        {
            var id = await MakeId();

            await _db.HashSetAsync(id, new[]
            {
                new HashEntry(ProviderField, metadata.Provider.ToString()),
                new HashEntry(TypeField, metadata.LinkType.ToString()),
                new HashEntry(IdField, metadata.Id)
            });

            await _db.KeyExpireAsync(id, TimeSpan.FromDays(1));

            return id;
        }

        public async Task<FeedMetadata> Load(string key)
        {
            if (string.IsNullOrWhiteSpace(key))
            {
                throw new ArgumentException("Feed key can't be empty");
            }

            var entries = await _db.HashGetAllAsync(key);

            if (entries.Length == 0)
            {
                throw new KeyNotFoundException("Invaid key");
            }

            var metadata = new FeedMetadata
            {
                Id = entries.Single(x => x.Name == IdField).Value,
                LinkType = ToEnum<LinkType>(entries.Single(x => x.Name == TypeField).Value),
                Provider = ToEnum<Provider>(entries.Single(x => x.Name == ProviderField).Value)
            };

            return metadata;
        }

        public Task ResetCounter()
        {
            return _db.KeyDeleteAsync(IdKey);
        }

        public async Task<string> MakeId()
        {
            var id = await _db.StringIncrementAsync(IdKey);
            return HashIds.EncodeLong(id);
        }

        private static T ToEnum<T>(string key)
        {
            return (T)Enum.Parse(typeof(T), key, true);
        }
    }
}