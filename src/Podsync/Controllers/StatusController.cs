using System;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Podsync.Services.Resolver;
using Podsync.Services.Storage;

namespace Podsync.Controllers
{
    public class StatusController : Controller
    {
        private const string ErrorStatus = "ERROR";

        private readonly IStorageService _storageService;
        private readonly IResolverService _resolverService;

        public StatusController(IStorageService storageService, IResolverService resolverService)
        {
            _storageService = storageService;
            _resolverService = resolverService;
        }

        public async Task<string> Index()
        {
            var storageStatus = ErrorStatus;

            try
            {
                var time = await _storageService.Ping();
                storageStatus = time.ToString();
            }
            catch (Exception)
            {
                // Nothing to do
            }

            var resolverStatus = _resolverService.Version ?? ErrorStatus;

            return $"Path: {Request.Path}\r\n" +
                   $"Redis: {storageStatus}\r\n" +
                   $"Resolve: {resolverStatus}";
        }
    }
}