using System.IO;
using System.Text;

namespace Podsync.Services.Rss.Feed.Internal
{
    public class Utf8StringWriter : StringWriter
    {
        public override Encoding Encoding { get; } = Encoding.UTF8;
    }
}