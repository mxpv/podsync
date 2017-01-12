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

        private const string Ytdl = "youtube-dl";

        private readonly ILogger _logger;

        public YtdlWrapper(IStorageService storageService, ILogger<YtdlWrapper> logger) : base(storageService)
        {
            _logger = logger;

            try
            {
                var cmd = Command.Run(Ytdl, "--version");
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


        protected override async Task<Uri> ResolveInternal(Uri videoUrl, ResolveFormat resolveFormat)
        {
            var format = SelectFormat(resolveFormat);

            try 
	        {	        
		        return await ResolveInternal(videoUrl, format);
	        }
	        catch (InvalidOperationException)
	        {
                // Give a try one more time, often it helps
	            await Task.Delay(WaitTimeoutBetweenFailedCalls);
                return await ResolveInternal(videoUrl, format);
            }
        }

        private static string SelectFormat(ResolveFormat format)
        {
            switch (format)
            {
                case ResolveFormat.VideoHigh:
                    return "best[ext=mp4]";
                case ResolveFormat.VideoLow:
                    return "worst[ext=mp4]";
                case ResolveFormat.AudioHigh:
                    return "bestaudio[ext=m4a]/worstaudio[ext=m4a]";
                case ResolveFormat.AudioLow:
                    return "worstaudio[ext=m4a]/bestaudio[ext=m4a]";
                default:
                    throw new ArgumentOutOfRangeException(nameof(format), "Unsupported format", null);
            }
        }

        private static IEnumerable<string> GetArguments(Uri videoUrl, string format)
        {
            // Video format code, see the "FORMAT SELECTION"
            yield return "-f";
            yield return format;

            // Simulate, quiet but print URL
            yield return "-g";
            yield return videoUrl.ToString();

            // Do not download the video and do not write anything to disk
            yield return "-s";

            // Suppress HTTPS certificate validation
            yield return "--no-check-certificate";

            // Do NOT contact the youtube-dl server for debugging
            yield return "--no-call-home";
        }

        private async Task<Uri> ResolveInternal(Uri videoUrl, string format)
        {
            var cmd = Command.Run(Ytdl, GetArguments(videoUrl, format), opts => opts.ThrowOnError().Timeout(ProcessWaitTimeout));

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