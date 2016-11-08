using System;
using System.Collections.Generic;
using System.Linq;
using System.Xml;
using System.Xml.Schema;
using System.Xml.Serialization;
using Shared;

namespace Podsync.Services.Feed
{
    public class Channel : IXmlSerializable
    {
        private const string PodsyncGeneratorName = "Podsync Generator";

        public Channel()
        {
            Items = Enumerable.Empty<Item>();
        }

        public string Title { get; set; }

        public string Description { get; set; }

        public string Link { get; set; }

        public DateTime LastBuildDate { get; set; }

        public DateTime PubDate { get; set; }

        public string Subtitle { get; set; }

        public string Summary { get; set; }

        public string Category { get; set; }

        public string Image { get; set; }

        public string Thumbnail { get; set; }

        public IEnumerable<Item> Items { get; set; }

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
            writer.WriteElementString("title", Title);
            writer.WriteElementString("description", Description);
            writer.WriteElementString("link", Link);

            writer.WriteElementString("generator", PodsyncGeneratorName);

            writer.WriteElementString("lastBuildDate", LastBuildDate.ToString("R"));
            writer.WriteElementString("pubDate", PubDate.ToString("R"));

            /*
                <itunes:subtitle>Laugh Factory</itunes:subtitle>
                <itunes:summary>The best stand up comedy clips online. That's it.</itunes:summary>
            */

            writer.WriteElementString("subtitle", Namespaces.Itunes, Title);
            writer.WriteElementString("summary", Namespaces.Itunes, Summary);

            /*
                <itunes:category text="TV & Film"/>
            */

            writer.WriteStartElement("category", Namespaces.Itunes);
            writer.WriteAttributeString("text", Category);
            writer.WriteEndElement();

            /*
                <itunes:image href="https://yt3.ggpht.com/photo.jpg"/>
            */

            writer.WriteStartElement("image", Namespaces.Itunes);
            writer.WriteAttributeString("href", Image);
            writer.WriteEndElement();

            /*
                <media:thumbnail url="https://yt3.ggpht.com//photo.jpg"/>
            */

            writer.WriteStartElement("thumbnail", Namespaces.Media);
            writer.WriteAttributeString("url", Thumbnail);
            writer.WriteEndElement();

            // Items
            var serializer = new XmlSerializer(typeof(Item));
            Items.ForEach(item => serializer.Serialize(writer, item));
        }
    }
}