using System;
using System.Linq;

namespace Shared
{
    internal static class StringExtensions
    {
        public static string TrimEnd(this string value, string word)
        {
            if (string.IsNullOrWhiteSpace(value) || string.IsNullOrWhiteSpace(word) || !value.EndsWith(word, StringComparison.OrdinalIgnoreCase))
            {
                return value;
            }

            return value.Substring(0, value.Length - word.Length);
        }

        public static string TrimStart(this string value, string word, StringComparison comparison = StringComparison.Ordinal)
        {
            if (string.IsNullOrWhiteSpace(value))
            {
                return value;
            }

            string result = value;
            while (result.StartsWith(word, comparison))
            {
                result = result.Substring(word.Length);
            }

            return result;
        }

        public static string RemoveChars(this string value, char[] list)
        {
            if (string.IsNullOrWhiteSpace(value))
            {
                return value;
            }

            return string.Concat(value.Split(list, StringSplitOptions.RemoveEmptyEntries));
        }

        public static string FirstCharToUpperInvariant(this string value)
        {
            if (string.IsNullOrWhiteSpace(value))
            {
                return value;
            }

            var first = value.First();
            if (char.IsUpper(first))
            {
                return value;
            }

            return first.ToString().ToUpperInvariant() + value.Substring(1);
        }
    }
}