using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Linq.Expressions;
using System.Net;
using System.Reflection;
using System.Threading;
using System.Threading.Tasks;
using HashidsNet;
using Microsoft.Extensions.Options;
using StackExchange.Redis;

namespace Podsync.Services.Storage
{
    public class RedisStorage : IStorageService
    {
        private const string CachePrefix = "cache";

        private const string IdKey = "keygen";
        private const string IdSalt = "65fce519433f4218aa0cee6394225eea";
        private const int IdLength = 4;

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

            var t = Db.CreateTransaction();
            t.AddCondition(Condition.KeyNotExists(id));

#pragma warning disable 4014
            // We should not await here because of transaction
            // See http://stackoverflow.com/questions/25976231/stackexchange-redis-transaction-methods-freezes
            t.HashSetAsync(id, BuildSet(metadata).ToArray());
            t.KeyExpireAsync(id, TimeSpan.FromDays(1));
#pragma warning restore 4014

            var succeeded = await t.ExecuteAsync();
            if (!succeeded)
            {
                throw new InvalidOperationException("Failed to save feed");
            }

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

            var metadata = new FeedMetadata();

            // V1
            UnpackProperty(metadata, x => x.Id, entries);
            UnpackProperty(metadata, x => x.Type, entries);
            UnpackProperty(metadata, x => x.Provider, entries);

            // V2
            UnpackProperty(metadata, x => x.Quality, entries, Constants.DefaultFormat);
            UnpackProperty(metadata, x => x.PageSize, entries, Constants.DefaultPageSize);
            UnpackProperty(metadata, x => x.PatreonId, entries, null);

            return metadata;
        }

        public Task Cache(string prefix, string id, string value, TimeSpan exp)
        {
            var key = BuildCacheKey(prefix, id);
            return Db.StringSetAsync(key, value, exp);
        }

        public async Task<string> GetCached(string prefix, string id)
        {
            var key = BuildCacheKey(prefix, id);
            var value = await Db.StringGetAsync(key);
            return value;
        }

        public async Task<string> MakeId()
        {
            var id = await Db.StringIncrementAsync(IdKey);
            return HashIds.EncodeLong(id);
        }

        private static string BuildCacheKey(string prefix, string id)
        {
            var key = $"{CachePrefix}:{prefix}:{id}";
            return key;
        }

        private static void UnpackProperty<T, P>(T target, Expression<Func<T, P>> memberLamda, HashEntry[] entries)
        {
            UnpackProperty(target, memberLamda, entries, default(P), true);
        }

        private static void UnpackProperty<T, P>(T target, Expression<Func<T, P>> memberLamda, HashEntry[] entries, P fallback, bool throwIfMissing = false)
        {
            var memberExpression = memberLamda.Body as MemberExpression;

            // Get property name via reflection
            var entryName = memberExpression?.Member?.Name;
            if (string.IsNullOrEmpty(entryName))
            {
                throw new InvalidOperationException("Wrong property expression");
            }

            P value;

            // RedisValue is value type
            if (entries.Any(x => string.Equals(x.Name, entryName, StringComparison.OrdinalIgnoreCase)))
            {
                var entry = entries.Single(x => string.Equals(x.Name, entryName, StringComparison.OrdinalIgnoreCase));

                var propertyType = typeof(P);
                if (propertyType.GetTypeInfo().IsEnum)
                {
                    value = (P)Enum.Parse(propertyType, entry.Value, true);
                }
                else
                {
                    value = (P)Convert.ChangeType(entry.Value, propertyType);
                }
            }
            else
            {
                if (throwIfMissing)
                {
                    throw new InvalidDataException("Missing mandatory property");
                }

                value = fallback;
            }

            var property = memberExpression.Member as PropertyInfo;
            property?.SetValue(target, value);
        }

        private IEnumerable<HashEntry> BuildSet(FeedMetadata metadata)
        {
            // V1.0
            yield return new HashEntry(nameof(metadata.Provider), metadata.Provider.ToString());
            yield return new HashEntry(nameof(metadata.Type), metadata.Type.ToString());
            yield return new HashEntry(nameof(metadata.Id), metadata.Id);

            // V2.0
            yield return new HashEntry(nameof(metadata.Quality), metadata.Quality.ToString());
            yield return new HashEntry(nameof(metadata.PageSize), metadata.PageSize);

            // V2.1
            if (!string.IsNullOrEmpty(metadata.PatreonId))
            {
                yield return new HashEntry(nameof(metadata.PatreonId), metadata.PatreonId);
            }
        }
    }
}