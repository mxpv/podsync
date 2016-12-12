using System;
using Microsoft.Extensions.DependencyInjection;

namespace Shared
{
    internal static class ServiceProviderExtensions
    {
        public static T CreateInstance<T>(this IServiceProvider serviceProvider)
        {
            return serviceProvider.CreateInstance<T>(typeof(T));
        }

        public static T CreateInstance<T>(this IServiceProvider serviceProvider, Type implementation)
        {
            return (T)serviceProvider.CreateInstance(implementation);
        }

        public static object CreateInstance(this IServiceProvider serviceProvider, Type type)
        {
            return ActivatorUtilities.CreateInstance(serviceProvider, type);
        }
    }
}