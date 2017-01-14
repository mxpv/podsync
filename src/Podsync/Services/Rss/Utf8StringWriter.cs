using System.IO;
using System.Text;

namespace Podsync.Services.Rss
{
    public class Utf8StringWriter : StringWriter
    {
        public override Encoding Encoding { get; } = Encoding.UTF8;
    }
}