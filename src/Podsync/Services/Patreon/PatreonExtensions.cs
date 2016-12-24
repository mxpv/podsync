using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Podsync.Services.Patreon.Contracts;

namespace Podsync.Services.Patreon
{
    public static class PatreonExtensions
    {
        public static async Task<User> FetchUserAndPledges(this IPatreonApi api, Tokens tokens)
        {
            var resp = await api.FetchProfile(tokens);

            dynamic userAttrs = resp.data.attributes;

            var user = new User
            {
                Id = resp.data.id,
                Email = userAttrs.email,
                Name = userAttrs.first_name ?? userAttrs.full_name,
                Url = userAttrs.url,
                Pledges = ParsePledges(resp)
            };

            return user;
        }

        private static IEnumerable<Pledge> ParsePledges(dynamic resp)
        {
            dynamic pledges = resp.data.relationships.pledges.data;

            foreach (var pledge in pledges)
            {
                var id = pledge.id;
                var type = pledge.type;

                foreach (var include in resp.included)
                {
                    if (include.id == id && include.type == type)
                    {
                        dynamic attrs = include.attributes;

                        yield return new Pledge
                        {
                            Id = include.id,
                            CreatedAt = ParseDate(attrs.created_at),
                            DeclinedSince = ParseDate(attrs.declined_since),
                            AmountCents = attrs.amount_cents,
                            PledgeCapCents = attrs.pledge_cap_cents
                        };

                        break;
                    }
                }
            }
        }

        private static DateTime ParseDate(object obj)
        {
            var date = obj?.ToString();

            if (string.IsNullOrWhiteSpace(date))
            {
                return DateTime.MinValue;
            }

            var dateTime = DateTime.Parse(date);

            return DateTime.SpecifyKind(dateTime, DateTimeKind.Utc);
        }
    }
}