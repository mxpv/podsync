# Schedule via cron expression

You can use `cron_schedule` field to build more precise update checks schedule.
A cron expression represents a set of times, using 5 space-separated fields.

| Field name   | Mandatory? | Allowed values  | Allowed special characters |
| ------------ | ---------- | --------------- | -------------------------- |
| Minutes      | Yes        | 0-59            | * / , -                    |
| Hours        | Yes        | 0-23            | * / , -                    |
| Day of month | Yes        | 1-31            | * / , - ?                  |
| Month        | Yes        | 1-12 or JAN-DEC | * / , -                    |
| Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?                  |

Month and Day-of-week field values are case insensitive. `SUN`, `Sun`, and `sun` are equally accepted.
The specific interpretation of the format is based on the Cron Wikipedia page: https://en.wikipedia.org/wiki/Cron

#### Predefined schedules

You may use one of several pre-defined schedules in place of a cron expression.

| Entry                   | Description                                | Equivalent to |
| ----------------------- | -------------------------------------------| ------------- |
| `@monthly`              | Run once a month, midnight, first of month | `0 0 1 * *`   |
| `@weekly`               | Run once a week, midnight between Sat/Sun  | `0 0 * * 0`   |
| `@daily (or @midnight)` | Run once a day, midnight                   | `0 0 * * *`   |
| `@hourly`               | Run once an hour, beginning of hour        | `0 * * * *`   |

#### Intervals

You may also schedule a job to execute at fixed intervals, starting at the time it's added
or cron is run. This is supported by formatting the cron spec like this:

    @every <duration>

where "duration" is a string accepted by [time.ParseDuration](http://golang.org/pkg/time/#ParseDuration).

For example, `@every 1h30m10s` would indicate a schedule that activates after 1 hour, 30 minutes, 10 seconds, and then every interval after that.
