using System;
using System.Collections.Generic;
using System.Linq;

namespace Podsync.Helpers
{
    internal static class EnumerableExtensions
    {
        public static IEnumerable<IEnumerable<T>> Chunk<T>(this IEnumerable<T> source, int chunkSize)
        {
            if (chunkSize < 1)
            {
                throw new ArgumentException("Chunk size must be positive", nameof(chunkSize));
            }

            while (source.Any())
            {
                yield return source.Take(chunkSize);
                source = source.Skip(chunkSize);
            }
        }

        public static bool SafeAny<T>(this IEnumerable<T> source)
        {
            if (source == null)
            {
                return false;
            }

            if (source.Any())
            {
                return true;
            }

            return false;
        }

        public static void ForEach<T>(this IEnumerable<T> source, Action<T> action)
        {
            if (source == null)
            {
                return;
            }

            foreach (var item in source)
            {
                action(item);
            }
        }

        public static void AddTo<T>(this IEnumerable<T> collection, List<T> target)
        {
            target.AddRange(collection);
        }

        public static void SafeForEach<T>(this IEnumerable<T> source, Action<T> action)
        {
            if (source == null)
            {
                return;
            }

            foreach (var item in source)
            {
                try
                {
                    action(item);
                }
                catch
                {
                    // Nothing to do
                }
            }
        }
    }
}