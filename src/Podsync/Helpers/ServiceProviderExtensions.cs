using System;
using System.Collections.Generic;
using System.Linq;
using System.Reflection;
using Microsoft.Extensions.DependencyInjection;

namespace Podsync.Helpers
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

        public static IEnumerable<Type> FindAllImplementationsOf(this IServiceProvider serviceProvider, Type interfaceType, Assembly assembly)
        {
            if (!interfaceType.GetTypeInfo().IsInterface)
            {
                throw new ArgumentException("T should be an interface");
            }

            return GetLoadableTypes(assembly).Where(type => IsAssignableFrom(interfaceType, type));
        }

        public static IEnumerable<Type> FindAllImplementationsOf<T>(this IServiceProvider serviceProvider, Assembly assembly)
        {
            return serviceProvider.FindAllImplementationsOf(typeof(T), assembly);
        }

        private static bool IsAssignableFrom(Type interfaceType, Type serviceType)
        {
            var serviceTypeInfo = serviceType.GetTypeInfo();
            if (serviceTypeInfo.IsInterface || serviceTypeInfo.IsAbstract)
            {
                return false;
            }

            var interfaceTypeInfo = interfaceType.GetTypeInfo();
            if (!interfaceTypeInfo.IsGenericType)
            {
                return interfaceType.IsAssignableFrom(serviceType);
            }

            return serviceType
                .GetInterfaces()
                .Where(type => type.GetTypeInfo().IsGenericType)
                .Any(type => type.GetGenericTypeDefinition() == interfaceType);
        }

        private static IEnumerable<Type> GetLoadableTypes(Assembly assembly)
        {
            try
            {
                return assembly.GetTypes();
            }
            catch (ReflectionTypeLoadException e)
            {
                return e.Types.Where(x => x != null);
            }
        }
    }
}