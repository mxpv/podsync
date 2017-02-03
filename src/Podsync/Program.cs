using System.IO;
using Microsoft.AspNetCore.Hosting;

namespace Podsync
{
    public class Program
    {
        public static void Main(string[] args)
        {
            var host = new WebHostBuilder()
                .UseKestrel(options =>
                {
                    options.AddServerHeader = false;

                    // Temporary workaround for Error -4047 EPIPE broken pipe
                    // Remove this after upgrade to 1.1.0
                    // See https://github.com/aspnet/KestrelHttpServer/issues/1182
                    options.ThreadCount = 1;
                })
                .UseContentRoot(Directory.GetCurrentDirectory())
                .UseIISIntegration()
                .UseStartup<Startup>()
                .Build();

            host.Run();
        }
    }
}
