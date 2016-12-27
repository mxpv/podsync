using System;
using System.Xml;
using System.Xml.Schema;
using System.Xml.Serialization;

namespace Podsync.Services.Feed
{
    [XmlRoot("item")]
    public class Item : IXmlSerializable
    {
        public string Title { get; set; }

        public string Description { get; set; }

        public string Author { get; set; }

        public Uri Link { get; set; }

        public DateTime PubDate { get; set; }

        public string Summary { get; set; }

        public TimeSpan Duration { get; set; }

        public MediaContent Content { get; set; }

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
            writer.WriteElementString("link", Link.ToString());
            writer.WriteElementString("pubDate", PubDate.ToString("R"));

            if (!string.IsNullOrWhiteSpace(Author))
            {
                writer.WriteElementString("author", Author);
            }

            /*
                <guid isPermaLink="true">https://youtube.com/watch?v=yp202t46OIE</guid>
            */

            writer.WriteStartElement("guid");
            writer.WriteAttributeString("isPermaLink", "true");
            writer.WriteString(Link.ToString());
            writer.WriteEndElement();

            /*
                <enclosure url="http://podsync.net/download/youtube/yp202t46OIE.mp4" length="48300000" type="video/mp4"/>
            */

            writer.WriteStartElement("enclosure");
            writer.WriteAttributeString("url", Content.Url.ToString());
            writer.WriteAttributeString("length", Content.Length.ToString());
            writer.WriteAttributeString("type", Content.MediaType);
            writer.WriteEndElement();

            /*
                <media:content url="http://podsync.net/download/youtube/yp202t46OIE.mp4" fileSize="48300000" type="video/mp4"/>
            */

            writer.WriteStartElement("content", Namespaces.Media);
            writer.WriteAttributeString("url", Content.Url.ToString());
            writer.WriteAttributeString("fileSize", Content.Length.ToString());
            writer.WriteAttributeString("type", Content.MediaType);
            writer.WriteEndElement();

            /*
                <itunes:subtitle>Mike E. Winfield - Cheating (Stand up comedy)</itunes:subtitle>
                <itunes:summary>...</itunes:summary>
                <itunes:duration>00:02:18</itunes:duration>
            */

            writer.WriteElementString("subtitle", Namespaces.Itunes, Title);
            writer.WriteElementString("summary", Namespaces.Itunes, Summary);
            writer.WriteElementString("duration", Namespaces.Itunes, Duration.ToString());
        }
    }
}