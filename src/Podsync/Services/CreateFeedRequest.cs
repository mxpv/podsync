using System.ComponentModel.DataAnnotations;
using Podsync.Services.Resolver;

namespace Podsync.Services
{
    public class CreateFeedRequest
    {
        [Required]
        [Url]
        public string Url { get; set; }

        public ResolveType? Quality { get; set; }

        [Range(50, 150)]
        public int? PageSize { get; set; }
    }
}