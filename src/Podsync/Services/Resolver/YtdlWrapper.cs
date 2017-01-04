using System;
using System.Diagnostics;
using System.IO;
using System.Threading.Tasks;

namespace Podsync.Services.Resolver
{
    public class YtdlWrapper : IResolverService
    {
        private static readonly int WaitTimeout = (int)TimeSpan.FromSeconds(5).TotalMilliseconds;

        public YtdlWrapper()
        {
            try
            {
                using (var proc = new Process())
                {
                    FillStartInfo(proc.StartInfo, "--version");

                    proc.Start();
                    proc.WaitForExit(WaitTimeout);

                    var stdout = proc.StandardOutput.ReadToEndAsync().GetAwaiter().GetResult();
                    Version = stdout;
                }
            }
            catch (Exception ex)
            {
                throw new FileNotFoundException("Failed to execute youtube-dl executable", "youtube-dl", ex);
            }
        }

        public string Version { get; }


        public async Task<Uri> Resolve(Uri videoUrl, ResolveType resolveType)
        {
            var format = SelectFormat(resolveType);

            try 
	        {	        
		        return await ResolveInternal(videoUrl, format);
	        }
	        catch (InvalidOperationException)
	        {
                // Give a try one more time, often it helps
	            await Task.Delay(TimeSpan.FromSeconds(1));
                return await ResolveInternal(videoUrl, format);
            }
        }

        private static void FillStartInfo(ProcessStartInfo startInfo, string arguments)
        {
            startInfo.FileName = "youtube-dl";
            startInfo.Arguments = arguments;

            startInfo.UseShellExecute = false;
            startInfo.CreateNoWindow = true;

            startInfo.RedirectStandardOutput = true;
            startInfo.RedirectStandardError = true;
        }

        private static string SelectFormat(ResolveType resolveType)
        {
            switch (resolveType)
            {
                case ResolveType.VideoHigh:
                    return "best[ext=mp4]/low[ext=mp4]";
                case ResolveType.VideoLow:
                    return "low[ext=mp4]/best[ext=mp4]";
                case ResolveType.AudioHigh:
                    return "bestaudio[ext=m4a]/worstaudio[ext=m4a]";
                case ResolveType.AudioLow:
                    return "worstaudio[ext=m4a]/bestaudio[ext=m4a]";
                default:
                    throw new ArgumentOutOfRangeException(nameof(resolveType), "Unsupported format", null);
            }
        }

        private static async Task<Uri> ResolveInternal(Uri videoUrl, string format)
        {
            using (var proc = new Process())
            {
                FillStartInfo(proc.StartInfo, $"-f {format} -g {videoUrl} --no-check-certificate");

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
                if (string.Equals(errout, "ERROR: requested format not available"))
                {
                    throw new NotSupportedException(errout);
                }

                throw new InvalidOperationException(errout);
            }
        }
    }
}