using System;
using Podsync.Services.Links;
using Xunit;

namespace Podsync.Tests.Services.Links
{
    public class LinkServiceTests : TestBase
    {
        private readonly ILinkService _linkService = new LinkService();

        [Theory]
        [InlineData("https://www.youtube.com/playlist?list=PLCB9F975ECF01953C", LinkType.Playlist, "PLCB9F975ECF01953C")]
        [InlineData("https://www.youtube.com/watch?v=otm9NaT9OWU&list=PLCB9F975ECF01953C", LinkType.Playlist, "PLCB9F975ECF01953C")]
        [InlineData("https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og", LinkType.Channel, "UC5XPnUk8Vvv_pWslhwom6Og")]
        [InlineData("https://www.youtube.com/user/UC5XPnUk8Vvv_pWslhwom6Og", LinkType.User, "UC5XPnUk8Vvv_pWslhwom6Og")]
        [InlineData("https://www.youtube.com/user/ComboBreakerVideo/videos", LinkType.User, "ComboBreakerVideo")]
        [InlineData("https://www.youtube.com/user/UC5XPnUk8Vvv_pWslhwom6Og/playlists", LinkType.User, "UC5XPnUk8Vvv_pWslhwom6Og")]
        [InlineData("https://www.youtube.com/playlist?list=PLP8qlV2aurYqdhyXW9ErqUW9Fw9F_mheM", LinkType.Playlist, "PLP8qlV2aurYqdhyXW9ErqUW9Fw9F_mheM")]
        [InlineData("https://www.youtube.com/user/NEMAGIA/videos", LinkType.User, "NEMAGIA")]
        public void ParseYoutubeLinks(string link, LinkType linkType, string id)
        {
            var info = _linkService.Parse(new Uri(link));

            Assert.Equal(info.Id, id);
            Assert.Equal(info.LinkType, linkType);
            Assert.Equal(info.Provider, Provider.YouTube);
        }

        [Theory]
        [InlineData("https://vimeo.com/groups/101", LinkType.Group, "101")]
        [InlineData("http://vimeo.com/groups/102", LinkType.Group, "102")]
        [InlineData("http://www.vimeo.com/groups/103", LinkType.Group, "103")]
        [InlineData("https://vimeo.com/awhitelabelproduct", LinkType.User, "awhitelabelproduct")]
        [InlineData("https://vimeo.com/groups/104/videos/", LinkType.Group, "104")]
        [InlineData("https://vimeo.com/channels/staffpicks", LinkType.Channel, "staffpicks")]
        [InlineData("https://vimeo.com/channels/staffpicks/146224925", LinkType.Channel, "staffpicks")]
        public void ParseVimeoLinks(string link, LinkType linkType, string id)
        {
            var info = _linkService.Parse(new Uri(link));

            Assert.Equal(info.Id, id);
            Assert.Equal(info.LinkType, linkType);
            Assert.Equal(info.Provider, Provider.Vimeo);
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
        [InlineData(Provider.Vimeo, LinkType.Channel, "yzx", "https://vimeo.com/channels/yzx")]
        [InlineData(Provider.Vimeo, LinkType.Group, "zyd", "https://vimeo.com/groups/zyd")]
        [InlineData(Provider.Vimeo, LinkType.User, "123", "https://vimeo.com/user123")]
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