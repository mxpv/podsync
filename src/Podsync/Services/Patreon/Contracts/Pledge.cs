using System;

namespace Podsync.Services.Patreon.Contracts
{
    public class Pledge
    {
        public string Id { get; set; }

        public DateTime CreatedAt { get; set; }

        public DateTime DeclinedSince { get; set; }

        public int AmountCents { get; set; }

        public int PledgeCapCents { get; set; }
    }
}