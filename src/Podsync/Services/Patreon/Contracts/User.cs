using System.Collections.Generic;
using System.Linq;

namespace Podsync.Services.Patreon.Contracts
{
    public class User
    {
        public User()
        {
            Pledges = Enumerable.Empty<Pledge>();
        }

        public string Id { get; set; }

        public string Email { get; set; }

        public string Name { get; set; }

        public string Url { get; set; }

        public IEnumerable<Pledge> Pledges { get; set; }
    }
}