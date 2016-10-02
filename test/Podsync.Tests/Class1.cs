using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using NuGet.Packaging;
using Xunit;

namespace Podsync.Tests
{
    public class Class1
    {
        public Class1()
        {
        }

        [Fact]
        public void PassingTest()
        {
            Assert.True(true);
        }

        [Fact]
        public void FailingTest()
        {
            Assert.False(true);
        }
    }
}
