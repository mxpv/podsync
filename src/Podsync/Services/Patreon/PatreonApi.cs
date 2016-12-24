using System;
using System.Net.Http;
using System.Net.Http.Headers;
using System.Threading.Tasks;
using Newtonsoft.Json.Linq;

namespace Podsync.Services.Patreon
{
    public sealed class PatreonApi : IPatreonApi
    {
        private readonly HttpClient _client;

        public PatreonApi()
        {
            _client = new HttpClient
            {
                BaseAddress = new Uri("https://api.patreon.com/oauth2/api/")
            };

            _client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer");
        }

        public Task<dynamic> FetchProfile(Tokens tokens)
        {
            return Query("current_user", tokens);
        }

        public void Dispose()
        {
            _client.Dispose();
        }

        private async Task<dynamic> Query(string path, Tokens tokens)
        {
            var request = new HttpRequestMessage(HttpMethod.Get, path);
            request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", tokens.AccessToken);

            var response = await _client.SendAsync(request);
            response.EnsureSuccessStatusCode();

            var json = await response.Content.ReadAsStringAsync();

            return JObject.Parse(json);
        }
    }
}