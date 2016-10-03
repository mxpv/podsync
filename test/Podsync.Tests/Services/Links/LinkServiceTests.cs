using System;
using Podsync.Services.Links;
using Xunit;

namespace Podsync.Tests.Services.Links
{
    public class LinkServiceTests
    {
        private readonly ILinkService _linkService = new LinkService();

        [Theory]
        [InlineData("http://youtu.be/jMeC7JFQ6801", Provider.YouTube, LinkType.Video, "jMeC7JFQ6801")]
        [InlineData("http://www.youtube.com/embed/watch?feature=player_embedded&v=jMeC7JFQ6802", Provider.YouTube, LinkType.Video, "jMeC7JFQ6802")]
        [InlineData("http://www.youtube.com/embed/watch?v=jMeC7JFQ6803", Provider.YouTube, LinkType.Video, "jMeC7JFQ6803")]
        [InlineData("http://www.youtube.com/embed/v=jMeC7JFQ6804", Provider.YouTube, LinkType.Video, "jMeC7JFQ6804")]
        [InlineData("http://www.youtube.com/watch?v=jMeC7JFQ6806", Provider.YouTube, LinkType.Video, "jMeC7JFQ6806")]
        [InlineData("http://www.youtube.com/watch?v=jMeC7JFQ6807", Provider.YouTube, LinkType.Video, "jMeC7JFQ6807")]
        [InlineData("http://www.youtu.be/jMeC7JFQ6808", Provider.YouTube, LinkType.Video, "jMeC7JFQ6808")]
        [InlineData("http://youtu.be/jMeC7JFQ6809", Provider.YouTube, LinkType.Video, "jMeC7JFQ6809")]
        [InlineData("http://www.youtube.com/watch?feature=player_embedded&v=jMeC7JFQ6805", Provider.YouTube, LinkType.Video, "jMeC7JFQ6805")]
        [InlineData("http://www.youtube.com/attribution_link?u=/watch?v=jMeC7JFQ6815&feature=share&a=9QlmP1yvjcllp0h3l0NwuA", Provider.YouTube, LinkType.Video, "jMeC7JFQ6815")]
        [InlineData("http://www.youtube.com/attribution_link?a=fF1CWYwxCQ4&u=/watch?v=jMeC7JFQ6816&feature=em-uploademail", Provider.YouTube, LinkType.Video, "jMeC7JFQ6816")]
        [InlineData("http://www.youtube.com/attribution_link?a=fF1CWYwxCQ4&feature=em-uploademail&u=/watch?v=jMeC7JFQ6817", Provider.YouTube, LinkType.Video, "jMeC7JFQ6817")]
        [InlineData("http://youtube.com/watch?v=jMeC7JFQ6810", Provider.YouTube, LinkType.Video, "jMeC7JFQ6810")]
        [InlineData("http://www.youtube.com/watch/jMeC7JFQ6811", Provider.YouTube, LinkType.Video, "jMeC7JFQ6811")]
        [InlineData("http://www.youtube.com/v/jMeC7JFQ6812", Provider.YouTube, LinkType.Video, "jMeC7JFQ6812")]
        [InlineData("http://WWW.YOUTUBE.COM/v/jMeC7JFQ6812", Provider.YouTube, LinkType.Video, "jMeC7JFQ6812")]
        [InlineData("https://www.youtube.com/playlist?list=PLCB9F975ECF01953C", Provider.YouTube, LinkType.Playlist, "PLCB9F975ECF01953C")]
        [InlineData("https://www.youtube.com/watch?v=otm9NaT9OWU&list=PLCB9F975ECF01953C", Provider.YouTube, LinkType.Playlist, "PLCB9F975ECF01953C")]
        [InlineData("https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og", Provider.YouTube, LinkType.Channel, "UC5XPnUk8Vvv_pWslhwom6Og")]
        [InlineData("https://www.youtube.com/user/UC5XPnUk8Vvv_pWslhwom6Og", Provider.YouTube, LinkType.User, "UC5XPnUk8Vvv_pWslhwom6Og")]
        [InlineData("https://www.youtube.com/user/ComboBreakerVideo/videos", Provider.YouTube, LinkType.User, "ComboBreakerVideo")]
        [InlineData("https://www.youtube.com/user/UC5XPnUk8Vvv_pWslhwom6Og/playlists", Provider.YouTube, LinkType.User, "UC5XPnUk8Vvv_pWslhwom6Og")]
        public void ParseLinkTest(string link, Provider provider, LinkType linkType, string id)
        {
            var info = _linkService.Parse(new Uri(link));

            Assert.Equal(info.Id, id);
            Assert.Equal(info.LinkType, linkType);
            Assert.Equal(info.Provider, provider);
        }

        [Fact]
        public void ParseInvalidLinkTest()
        {
            Assert.Throws<ArgumentNullException>(() => _linkService.Parse(null));
            Assert.Throws<ArgumentException>(() => _linkService.Parse(new Uri("http://www.apple.com")));
        }

        [Theory]
        [InlineData(Provider.YouTube, LinkType.Channel, "123", "https://youtube.com/channel/123")]
        [InlineData(Provider.YouTube, LinkType.Playlist, "213", "https://youtube.com/playlist?list=213")]
        [InlineData(Provider.YouTube, LinkType.Video, "321", "https://youtube.com/watch?v=321")]
        [InlineData(Provider.YouTube, LinkType.Info, "111", "https://youtube.com/get_video_info?video_id=111")]
        [InlineData(Provider.Vimeo, LinkType.Category, "xyz", "https://vimeo.com/categories/xyz")]
        [InlineData(Provider.Vimeo, LinkType.Channel, "yzx", "https://vimeo.com/channels/yzx")]
        [InlineData(Provider.Vimeo, LinkType.Group, "zyd", "https://vimeo.com/groups/zyd")]
        [InlineData(Provider.Vimeo, LinkType.User, "dfz", "https://vimeo.com/dfz")]
        [InlineData(Provider.Vimeo, LinkType.Info, "xgd", "https://player.vimeo.com/video/xgd/config")]
        public void MakeLinkTest(Provider provider, LinkType linkType, string id, string expected)
        {
            var info = new LinkInfo
            {
                Id = id,
                LinkType = linkType,
                Provider = provider,
            };

            var link = _linkService.Make(info);
            Assert.Equal(expected, link.ToString());
        }
    }
}