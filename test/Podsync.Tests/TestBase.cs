using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Options;
using Podsync.Services;

namespace Podsync.Tests
{
    public abstract class TestBase
    {
        private const string UserSecretsId = "aspnet-Podsync-20161004104901";

        protected TestBase()
        {
            var configurationRoot = new ConfigurationBuilder()
                .AddUserSecrets(UserSecretsId)
                .Build();

            var podsyncSection = configurationRoot.GetSection("Podsync");

            var configuration = new PodsyncConfiguration();
            podsyncSection.Bind(configuration);

            Options = new OptionsWrapper<PodsyncConfiguration>(configuration);
        }

        protected IOptions<PodsyncConfiguration> Options { get; }

        protected PodsyncConfiguration Configuration => Options.Value;        
    }
}