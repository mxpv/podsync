# Podsync

![Podsync](docs/img/logo.png)

[![](https://github.com/mxpv/podsync/workflows/CI/badge.svg)](https://github.com/mxpv/podsync/actions?query=workflow%3ACI)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/mxpv/podsync)](https://github.com/mxpv/podsync/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/mxpv/podsync)](https://goreportcard.com/report/github.com/mxpv/podsync)
[![Patreon](https://img.shields.io/badge/support-patreon-E6461A.svg)](https://www.patreon.com/podsync)
[![Twitter Follow](https://img.shields.io/twitter/follow/pod_sync?style=social)](https://twitter.com/pod_sync)

Podsync - is a simple, free service that lets you listen to any YouTube / Vimeo channels, playlists or user videos in
podcast format.

Podcast applications have a rich functionality for content delivery - automatic download of new episodes,
remembering last played position, sync between devices and offline listening. This functionality is not available
on YouTube and Vimeo. So the aim of Podsync is to make your life easier and enable you to view/listen to content on
any device in podcast client.

## Features

- Works with YouTube and Vimeo.
- Supports feeds configuration: video/audio, high/low quality, max video height, etc.
- mp3 encoding
- Update scheduler supports cron expressions
- Episodes filtering (match by title).
- Feeds customizations (custom artwork, category, language, etc).
- Supports episodes cleanup (keep last X episodes).
- One-click deployment for AWS.
- Runs on Windows, Mac OS, Linux, and Docker.
- Supports ARM.

## Dependencies

If you're running the CLI as binary (e.g. not via Docker), you need to make sure that dependencies are available on
your system. Currently Podsync depends on `youtube-dl` and `ffmpeg`.

On Mac you can install those with `brew`:
```
brew install youtube-dl ffmpeg
```

## Access tokens

In order to query YouTube or Vimeo API you have to obtain an API token first.

- [How to get YouTube API key](https://elfsight.com/blog/2016/12/how-to-get-youtube-api-key-tutorial/)
- [Generate an access token for Vimeo](https://developer.vimeo.com/api/guides/start#generate-access-token)

## Configuration example

You need to create a configuration file (for instance `config.toml`) and specify the list of feeds that you're going to host.
Here is an example how configuration might look like:

```toml
[server]
port = 8080
data_dir = "/app/data" # Don't change if you run podsync via docker

[tokens]
youtube = "{YOUTUBE_API_TOKEN}" # Tokens from `Access tokens` section
vimeo = "{VIMEO_API_TOKEN}"

[feeds]
  [feeds.ID1]
  url = "{FEED_URL}" # URL address of a channel, group, user, or playlist. 
  page_size = 50 # The number of episodes to query each update (keep in mind, that this might drain API token)
  update_period = "12h" # How often query for updates, examples: "60m", "4h", "2h45m"
  quality = "high" # or "low"
  format = "video" # or "audio"
  # cover_art = "{IMAGE_URL}" # Optional URL address of an image file
  # max_height = "720" # Optional maximal height of video, example: 720, 1080, 1440, 2160, ...
  # cron_schedule = "@every 12h" # Optional cron expression format. If set then overwrite 'update_period'. See details below
  # filters = { title = "regex for title here" } # Optional Golang regexp format. If set, then only download episodes with matching titles.
  # clean = { keep_last = 10 } # Keep last 10 episodes (order desc by PubDate)
```

Episodes files will be kept at: `/path/to/data/directory/ID1`, feed will be accessible from: `http://localhost/ID1.xml`

If you want to hide Podsync behind reverse proxy like nginx, you can use `hostname` field:

```toml
[server]
port = 8080
hostname = "https://my.test.host:4443"

[feeds]
  [feeds.ID1]
  ...
```

Server will be accessible from `http://localhost:8080`, but episode links will point to `https://my.test.host:4443/ID1/...`


### Schedule via cron expression

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

## One click deployment

[![Deploy to AWS](https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/new?stackName=Podsync&templateURL=https://podsync-cf.s3.amazonaws.com/cloud_formation.yml)

## How to run

### Run as binary:
```
$ ./podsync --config config.toml
```

### Run via Docker:
```
$ docker pull mxpv/podsync:latest
$ docker run \
    -p 8080:8080 \
    -v $(pwd)/data:/app/data/ \
    -v $(pwd)/config.toml:/app/config.toml \
    mxpv/podsync:latest
```

### Run via Docker Compose:
```
$ docker-compose up
```

## How to make a release

Just push a git tag. CI will do the rest.

