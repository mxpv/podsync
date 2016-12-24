using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Threading;
using System.Threading.Tasks;
using HashidsNet;
using Microsoft.Extensions.Options;
using Podsync.Services.Links;
using Podsync.Services.Resolver;
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
        private const string QualityField = "quality";
        private const string PageSizeField = "pageSize";

        private const ResolveType DefaultQuality = ResolveType.VideoHigh;
        private const int DefaultPageSize = 50;

        private static readonly IHashids HashIds = new Hashids(IdSalt, IdLength);

        private readonly string _cs;
        private IDatabase _db;

        public RedisStorage(IOptions<PodsyncConfiguration> configuration)
        {
            _cs = configuration.Value.RedisConnectionString;
        }

        private IDatabase Db
        {
            get { return LazyInitializer.EnsureInitialized(ref _db, () => Connect(_cs).GetAwaiter().GetResult().GetDatabase()); }
        }

        private static async Task<ConnectionMultiplexer> Connect(string cs)
        {
            var options = ConfigurationOptions.Parse(cs);

            try
            {
                return await ConnectionMultiplexer.ConnectAsync(options);
            }
            catch (PlatformNotSupportedException)
            {
                // Can't connect via address on Linux environments, using workaround
                // See https://github.com/StackExchange/StackExchange.Redis/issues/410#issuecomment-246332140
                var addressEndpoint = options.EndPoints.SingleOrDefault() as DnsEndPoint;
                if (addressEndpoint != null)
                {
                    var ip = await Dns.GetHostEntryAsync(addressEndpoint.Host);
                    options.EndPoints.Remove(addressEndpoint);
                    options.EndPoints.Add(ip.AddressList.First(), addressEndpoint.Port);
                }

                return await ConnectionMultiplexer.ConnectAsync(options);
            }
        }

        public void Dispose()
        {
            if (_db != null)
            {
                _db.Multiplexer.Dispose();
                _db = null;
            }
        }

        public Task<TimeSpan> Ping()
        {
            return Db.PingAsync();
        }

        public async Task<string> Save(FeedMetadata metadata)
        {
            var id = await MakeId();

            await Db.HashSetAsync(id, new[]
            {
                new HashEntry(ProviderField, metadata.Provider.ToString()),
                new HashEntry(TypeField, metadata.LinkType.ToString()),
                new HashEntry(IdField, metadata.Id),
                new HashEntry(QualityField, metadata.Quality.ToString()),
                new HashEntry(PageSizeField, metadata.PageSize), 
            });

            await Db.KeyExpireAsync(id, TimeSpan.FromDays(1));

            return id;
        }

        public async Task<FeedMetadata> Load(string key)
        {
            if (string.IsNullOrWhiteSpace(key))
            {
                throw new ArgumentException("Feed key can't be empty");
            }

            var entries = await Db.HashGetAllAsync(key);

            // Expire after 3 month if no use
            await Db.KeyExpireAsync(key, TimeSpan.FromDays(90));

            if (entries.Length == 0)
            {
                throw new KeyNotFoundException("Invaid key");
            }

            var metadata = new FeedMetadata
            {
                Id = entries.Single(x => x.Name == IdField).Value,
                LinkType = ToEnum<LinkType>(entries.Single(x => x.Name == TypeField)),
                Provider = ToEnum<Provider>(entries.Single(x => x.Name == ProviderField)),
            };

            if (entries.Length > 3)
            {
                metadata.Quality = ToEnum<ResolveType>(entries.Single(x => x.Name == QualityField));
                metadata.PageSize = (int)entries.Single(x => x.Name == PageSizeField).Value;
            }
            else
            {
                // Set default values
                metadata.Quality = DefaultQuality;
                metadata.PageSize = DefaultPageSize;
            }

            return metadata;
        }

        public Task ResetCounter()
        {
            return Db.KeyDeleteAsync(IdKey);
        }

        public async Task<string> MakeId()
        {
            var id = await Db.StringIncrementAsync(IdKey);
            return HashIds.EncodeLong(id);
        }

        private static T ToEnum<T>(HashEntry key)
        {
            return (T)Enum.Parse(typeof(T), key.Value, true);
        }
    }
}