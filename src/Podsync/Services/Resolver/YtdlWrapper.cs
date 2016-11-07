using System;
using System.Diagnostics;
using System.Threading.Tasks;

namespace Podsync.Services.Resolver
{
    public class YtdlWrapper : IResolverService
    {
        private static readonly int WaitTimeout = (int)TimeSpan.FromSeconds(5).TotalMilliseconds;

        public async Task<Uri> Resolve(Uri videoUrl, FileType fileType, Quality quality)
        {
            var format = SelectFormat(fileType, quality);

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

        private static string SelectFormat(FileType fileType, Quality quality)
        {
            if (fileType == FileType.Video)
            {
                if (quality == Quality.High)
                {
                    return "bestvideo";
                }

                if (quality == Quality.Low)
                {
                    return "worstvideo";
                }
            }
            else if (fileType == FileType.Audio)
            {
                if (quality == Quality.High)
                {
                    return "bestaudio";
                }

                if (quality == Quality.Low)
                {
                    return "worstaudio";
                }
            }

            throw new ArgumentException("Unsupported format");
        }
    }
}