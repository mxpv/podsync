using System;
using System.Net;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;
using Microsoft.Extensions.Logging;
using Podsync.Services;

namespace Podsync.Helpers
{
    public class HandleExceptionAttribute : ExceptionFilterAttribute
    {
        private readonly ILogger _logger;

        public HandleExceptionAttribute(ILogger<HandleExceptionAttribute> logger)
        {
            _logger = logger;
        }

        public override void OnException(ExceptionContext context)
        {
            var exception = context.Exception;
            if (exception is ArgumentNullException || exception is ArgumentException)
            {
                context.Result = new BadRequestObjectResult(exception.Message);
            }
            else
            {
                context.Result = new StatusCodeResult((int)HttpStatusCode.InternalServerError);

                _logger.LogCritical(Constants.Events.UnhandledError, context.Exception, "Unhandled exception");
            }
        }
    }
}