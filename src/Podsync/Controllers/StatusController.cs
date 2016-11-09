using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Podsync.Services.Storage;

namespace Podsync.Controllers
{
    public class StatusController : Controller
    {
        private readonly IStorageService _storageService;

        public StatusController(IStorageService storageService)
        {
            _storageService = storageService;
        }

        public async Task<string> Index()
        {
            var time = await _storageService.Ping();

            return $"Path: {Request.Path}\r\n" +
                   $"Redis: {time}";
        }
    }
}