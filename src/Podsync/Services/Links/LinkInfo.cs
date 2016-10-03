namespace Podsync.Services.Links
{
    public struct LinkInfo
    {
        public string Id { get; set; }

        public LinkType LinkType { get; set; }

        public Provider Provider { get; set; }
    }
}