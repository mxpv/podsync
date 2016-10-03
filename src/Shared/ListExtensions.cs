using System.Collections.Generic;

namespace Shared
{
    internal static class ListExtensions
    {
        public static void AddRange<T>(this IList<T> source, T[] elements)
        {
            if (source == null || elements == null)
            {
                return;
            }

            foreach (var element in elements)
            {
                source.Add(element);
            }
        }
    }
}