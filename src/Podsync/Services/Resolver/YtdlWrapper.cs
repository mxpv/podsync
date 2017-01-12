using System;
using System.Collections.Generic;
using System.IO;
using System.Threading.Tasks;
using Medallion.Shell;
using Microsoft.Extensions.Logging;
using Podsync.Services.Storage;

namespace Podsync.Services.Resolver
{
    public class YtdlWrapper : CachedResolver
    {
        private static readonly TimeSpan ProcessWaitTimeout = TimeSpan.FromMinutes(1);
        private static readonly TimeSpan WaitTimeoutBetweenFailedCalls = TimeSpan.FromSeconds(30);
        
        private const string YtdlName = "youtube-dl";
        
        private readonly ILogger _logger;

        public YtdlWrapper(IStorageService storageService, ILogger<YtdlWrapper> logger) : base(storageService)
        {
            _logger = logger;

            try
            {
                var cmd = Command.Run(YtdlName, "--version");
                var version = cmd.Result.StandardOutput;

                Version = version;

                _logger.LogInformation("Uring youtube-dl {VERSION}", version);
            }
            catch (Exception ex)
            {
                throw new FileNotFoundException("Failed to execute youtube-dl executable", "youtube-dl", ex);
            }
        }

        public override string Version { get; }


        protected override async Task<Uri> ResolveInternal(Uri videoUrl, ResolveFormat format)
        {
            try 
	        {	        
		        return await Ytdl(videoUrl, format);
	        }
	        catch (InvalidOperationException)
	        {
                // Give a try one more time, often it helps
	            await Task.Delay(WaitTimeoutBetweenFailedCalls);
                return await Ytdl(videoUrl, format);
            }
        }

        private static IEnumerable<string> GetArguments(Uri videoUrl, ResolveFormat format)
        {
            var host = videoUrl.Host.ToLowerInvariant();

            // Video format code, see the "FORMAT SELECTION"
            yield return "-f";

            if (host.Contains("youtube.com"))
            {
                if (format == ResolveFormat.VideoHigh)
                {
                    yield return "best[ext=mp4]";
                }
                else if (format == ResolveFormat.VideoLow)
                {
                    yield return "worst[ext=mp4]";
                }
                else if (format == ResolveFormat.AudioHigh)
                {
                    yield return "bestaudio[ext=m4a]/worstaudio[ext=m4a]";
                }
                else if (format == ResolveFormat.AudioLow)
                {
                    yield return "worstaudio[ext=m4a]/bestaudio[ext=m4a]";
                }
                else
                {
                    throw new ArgumentException("Unsupported resolve format");
                }
            }
            else if (host.Contains("vimeo.com"))
            {
                if (format == ResolveFormat.VideoHigh)
                {
                    yield return "Original/http-1080p/http-720p/http-360p/http-270p";
                }
                else if (format == ResolveFormat.VideoLow)
                {
                    yield return "http-270p/http-360p/http-540p/http-720p/http-1080p";
                }
                else
                {
                    throw new ArgumentException("Unsupported resolve format");
                }
            }
            else
            {
                throw new ArgumentException("Unsupported video provider");
            }

            // Simulate, quiet but print URL
            yield return "-g";

            // Do not download the video and do not write anything to disk
            yield return "-s";

            // Suppress HTTPS certificate validation
            yield return "--no-check-certificate";

            // Do NOT contact the youtube-dl server for debugging
            yield return "--no-call-home";

            yield return videoUrl.ToString();
        }

        private async Task<Uri> Ytdl(Uri videoUrl, ResolveFormat format)
        {
            var cmd = Command.Run(YtdlName, GetArguments(videoUrl, format), opts => opts.ThrowOnError().Timeout(ProcessWaitTimeout));

            try
            {
                await cmd.Task;
            }
            catch (ErrorExitCodeException ex)
            {
                var errout = await cmd.StandardError.ReadToEndAsync();
                var msg = !string.IsNullOrWhiteSpace(errout) ? errout : ex.Message;

                _logger.LogError(Constants.Events.YtdlError, ex, "Failed to resolve {URL} in format {FORMAT}", videoUrl, format);

                if (string.Equals(errout, "ERROR: requested format not available"))
                {
                    throw new NotSupportedException("Requested format not available", ex);
                }

                throw new InvalidOperationException(msg, ex);
            }

            var stdout = await cmd.StandardOutput.ReadToEndAsync();
            if (Uri.IsWellFormedUriString(stdout, UriKind.Absolute))
            {
                return new Uri(stdout);
            }

            throw new InvalidOperationException(stdout);
        }
    }
}