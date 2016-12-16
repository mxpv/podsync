using System.Linq;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;

namespace Podsync.Helpers
{
    /// <summary>
    /// Validate the model state prior to action execution.
    /// See http://www.strathweb.com/2015/06/action-filters-service-filters-type-filters-asp-net-5-mvc-6/
    /// </summary>
    public class ValidateModelStateAttribute : ActionFilterAttribute
    {
        private readonly bool _validateNotNull;

        public ValidateModelStateAttribute(bool validateNotNull = true)
        {
            _validateNotNull = validateNotNull;
        }

        public override void OnActionExecuting(ActionExecutingContext context)
        {
            if (!context.ModelState.IsValid || _validateNotNull && context.ActionArguments.Any(p => p.Value == null))
            {
                context.Result = new BadRequestObjectResult(context.ModelState);
            }
        }
    }
}