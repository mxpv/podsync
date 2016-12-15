using Microsoft.AspNetCore.Mvc;

namespace Podsync.Controllers
{
    public class HomeController : Controller
    {
        public IActionResult Index()
        {
            return View();
        }
    }
}
