using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Podsync.Services.Resolver;
using Podsync.Services.Storage;

namespace Podsync.Controllers
{
    public class StatusController : Controller
    {
        private readonly IStorageService _storageService;
        private readonly IResolverService _resolverService;

        public StatusController(IStorageService storageService, IResolverService resolverService)
        {
            _storageService = storageService;
            _resolverService = resolverService;
        }

        public async Task<string> Index()
        {
            var time = await _storageService.Ping();

            return $"Path: {Request.Path}\r\n" +
                   $"Redis: {time}\r\n" +
                   $"Resolve: {_resolverService.Version}";
        }
    }
}