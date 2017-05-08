using System;
using System.Collections.Generic;
using System.Linq;
using System.Xml;
using System.Xml.Schema;
using System.Xml.Serialization;
using Podsync.Helpers;

namespace Podsync.Services.Rss.Contracts
{
    [XmlRoot("rss")]
    public class Feed : IXmlSerializable
    {
        public const string Version = "2.0";

        public IEnumerable<Channel> Channels { get; set; }

        public Feed()
        {
            Channels = Enumerable.Empty<Channel>();
        }

        public XmlSchema GetSchema()
        {
            return null;
        }

        public void ReadXml(XmlReader reader)
        {
            throw new NotSupportedException("Reading is not supported");
        }

        public void WriteXml(XmlWriter writer)
        {
            writer.WriteAttributeString("xmlns", "dc", null, Namespaces.Dc);
            writer.WriteAttributeString("xmlns", "content", null, Namespaces.Content);
            writer.WriteAttributeString("xmlns", "atom", null, Namespaces.Atom);
            writer.WriteAttributeString("xmlns", "itunes", null, Namespaces.Itunes);
            writer.WriteAttributeString("xmlns", "media", null, Namespaces.Media);

            writer.WriteAttributeString("version", Version);

            var serializer = new XmlSerializer(typeof(Channel));
            Channels.ForEach(channel => serializer.Serialize(writer, channel));
        }

        public override string ToString()
        {
            var serializer = new XmlSerializer(typeof(Feed));

            // Serialize feed to XML string
            using (var writer = new Utf8StringWriter())
            {
                serializer.Serialize(writer, this);
                return writer.ToString();
            }
        }
    }
}