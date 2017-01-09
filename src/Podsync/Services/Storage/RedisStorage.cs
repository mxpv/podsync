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

            if (await Db.KeyExistsAsync(id))
            {
                throw new InvalidOperationException("Failed to generate feed id");
            }

            await Db.HashSetAsync(id, new[]
            {
                // V1
                new HashEntry(nameof(metadata.Provider), metadata.Provider.ToString()),
                new HashEntry(nameof(metadata.Type), metadata.Type.ToString()),
                new HashEntry(nameof(metadata.Id), metadata.Id),

                // V2
                new HashEntry(nameof(metadata.Quality), metadata.Quality.ToString()),
                new HashEntry(nameof(metadata.PageSize), metadata.PageSize),
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

            var metadata = new FeedMetadata();

            // V1
            SetProperty(metadata, x => x.Id, entries);
            SetProperty(metadata, x => x.Type, entries);
            SetProperty(metadata, x => x.Provider, entries);

            // V2
            SetProperty(metadata, x => x.Quality, entries, Constants.DefaultFormat);
            SetProperty(metadata, x => x.PageSize, entries, Constants.DefaultPageSize);

            return metadata;
        }

        public async Task<string> MakeId()
        {
            var id = await Db.StringIncrementAsync(IdKey);
            return HashIds.EncodeLong(id);
        }

        private static void SetProperty<T, P>(T target, Expression<Func<T, P>> memberLamda, HashEntry[] entries)
        {
            SetProperty(target, memberLamda, entries, default(P), true);
        }

        private static void SetProperty<T, P>(T target, Expression<Func<T, P>> memberLamda, HashEntry[] entries, P fallback, bool throwIfMissing = false)
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
                    value = (P)Enum.Parse(propertyType, entry.Value);
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
    }
}