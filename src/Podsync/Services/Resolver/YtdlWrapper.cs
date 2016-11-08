using System;
using System.Diagnostics;
using System.Threading.Tasks;

namespace Podsync.Services.Resolver
{
    public class YtdlWrapper : IResolverService
    {
        private static readonly int WaitTimeout = (int)TimeSpan.FromSeconds(5).TotalMilliseconds;

        public async Task<Uri> Resolve(Uri videoUrl, ResolveType resolveType)
        {
            var format = SelectFormat(resolveType);

            using (var proc = new Process())
            {
                var startInfo = proc.StartInfo;

                startInfo.FileName = "youtube-dl";
                startInfo.Arguments = $"-f {format} -g {videoUrl}";

                startInfo.UseShellExecute = false;
                startInfo.CreateNoWindow = true;

                startInfo.RedirectStandardOutput = true;
                startInfo.RedirectStandardError = true;

                proc.Start();

                if (!proc.WaitForExit(WaitTimeout))
                {
                    proc.Kill();

                    throw new InvalidOperationException("Can't resolve URL because of timeout");
                }

                var stdout = await proc.StandardOutput.ReadToEndAsync();
                if (Uri.IsWellFormedUriString(stdout, UriKind.Absolute))
                {
                    return new Uri(stdout);
                }

                var errout = await proc.StandardError.ReadToEndAsync();
                throw new InvalidOperationException(errout);
            }
        }

        private static string SelectFormat(ResolveType resolveType)
        {
            switch (resolveType)
            {
                case ResolveType.VideoHigh:
                    return "best";
                case ResolveType.VideoLow:
                    return "worst";
                case ResolveType.AudioHigh:
                    return "bestaudio";
                case ResolveType.AudioLow:
                    return "worstaudio";
                default:
                    throw new ArgumentOutOfRangeException(nameof(resolveType), "Unsupported format", null);
            }
        }
    }
}