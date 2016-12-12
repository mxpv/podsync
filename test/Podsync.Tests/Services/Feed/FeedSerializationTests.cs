using System;
using System.IO;
using System.Xml.Serialization;
using Podsync.Services.Feed;
using Xunit;

namespace Podsync.Tests.Services.Feed
{
    public class FeedSerializationTests
    {
        private static readonly Uri ImageUrl = new Uri("https://yt3.ggpht.com/-OOqFwHRjeMQ/AAAAAAAAAAI/AAAAAAAAAAA/0XQZ2NeGp_0/s88-c-k-no-mo-rj-c0xffffff/photo.jpg");

        [Fact]
        public void SerializeFeedTest()
        {
            var feed = new Rss();

            var item = new Item
            {
                Title = "Steve Gillespie - Getting Arrested (Stand up Comedy)",
                Link = new Uri("https://youtube.com/watch?v=Jj22gfTnpAI"),
                PubDate = DateTime.Parse("Mon, 07 Nov 2016 20:02:26 GMT"),
                Content = new MediaContent
                {
                    Url = new Uri("http://podsync.net/download/youtube/Jj22gfTnpAI.mp4"),
                    Length = 52850000,
                    MediaType = "video/mp4"
                },
                Duration = new TimeSpan(0, 0, 2, 31)
            };

            var channel = new Channel
            {
                Title = "Laugh Factory",
                Description = "The best stand up comedy clips online. That's it.",
                Link = new Uri("https://youtube.com/channel/UCxyCzPY2pjAjrxoSYclpuLg"),
                LastBuildDate = DateTime.Parse("Tue, 08 Nov 2016 05:55:25 GMT"),
                PubDate = DateTime.Parse("Mon, 31 Jul 2006 22:18:05 GMT"),
                Subtitle = "Laugh Factory",
                Summary = "The best stand up comedy clips online. That's it.",
                Category = "TV & Film",
                Image = ImageUrl,
                Thumbnail = ImageUrl,

                Items = new[]
                {
                    item
                }
            };

            feed.Channels = new[]
            {
                channel
            };

            var serializer = new XmlSerializer(typeof(Rss));

            string body;
            using (var writer = new StringWriter())
            {
                serializer.Serialize(writer, feed);
                body = writer.ToString();
            }

            Assert.NotEmpty(body);

            // Channel tests

            Assert.Contains("<title>Laugh Factory</title>", body);
            Assert.Contains("<description>The best stand up comedy clips online. That's it.</description>", body);
            Assert.Contains("<link>https://youtube.com/channel/UCxyCzPY2pjAjrxoSYclpuLg</link>", body);
            Assert.Contains("<generator>Podsync Generator</generator>", body);

            Assert.Contains("<itunes:subtitle>Laugh Factory</itunes:subtitle>", body);
            Assert.Contains("<itunes:summary>The best stand up comedy clips online. That's it.</itunes:summary>", body);
            Assert.Contains("<itunes:category text=\"TV &amp; Film\" />", body);
            Assert.Contains($"<itunes:image href=\"{ImageUrl}\" />", body);
            Assert.Contains($"<media:thumbnail url=\"{ImageUrl}\" />", body);

            // Items tests
            Assert.Contains("<title>Steve Gillespie - Getting Arrested (Stand up Comedy)</title>", body);
            Assert.Contains("<itunes:subtitle>Steve Gillespie - Getting Arrested (Stand up Comedy)</itunes:subtitle>", body);

            Assert.Contains("<link>https://youtube.com/watch?v=Jj22gfTnpAI</link>", body);
            Assert.Contains("<guid isPermaLink=\"true\">https://youtube.com/watch?v=Jj22gfTnpAI</guid>", body);

            Assert.Contains("<itunes:duration>00:02:31</itunes:duration>", body);

            Assert.Contains("<enclosure url=\"http://podsync.net/download/youtube/Jj22gfTnpAI.mp4\" length=\"52850000\" type=\"video/mp4\" />", body);
            Assert.Contains("<media:content url=\"http://podsync.net/download/youtube/Jj22gfTnpAI.mp4\" fileSize=\"52850000\" type=\"video/mp4\" />", body);
        }
    }
}