using System;
using System.Net;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;

namespace Podsync.Helpers
{
    public class HandleExceptionAttribute : ExceptionFilterAttribute
    {
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
            }
        }
    }
}