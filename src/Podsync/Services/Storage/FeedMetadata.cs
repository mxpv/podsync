﻿using Podsync.Services.Links;
using Podsync.Services.Resolver;

namespace Podsync.Services.Storage
{
    public class FeedMetadata
    {
        public Provider Provider { get; set; }

        public LinkType Type { get; set; }

        public string Id { get; set; }

        public ResolveFormat Quality { get; set; }

        public int PageSize { get; set; }

        /// <summary>
        /// User id on Patreon
        /// </summary>
        public string PatreonId { get; set; }

        public override string ToString() => $"{Provider} ({LinkType}) {Id}";

        // Workaround for backward compatibility
        public LinkType LinkType
        {
            get { return Type; }
            set { Type = value; }
        }
    }
}