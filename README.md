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
- OPML export.
- Supports episodes cleanup (keep last X episodes).
- One-click deployment for AWS.
- Runs on Windows, Mac OS, Linux, and Docker.
- Supports ARM.
- Automatic youtube-dl self update.
- Supports API keys rotation.

## Dependencies

If you're running the CLI as binary (e.g. not via Docker), you need to make sure that dependencies are available on
your system. Currently, Podsync depends on `youtube-dl` and `ffmpeg`.

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
# Bind a specific IP addresses for server ,"*": bind all IP addresses which is default option, localhost or 127.0.0.1  bind a single IPv4 address
bind_address = "172.20.10.2" 
# Specify path for reverse proxy and only [A-Za-z0-9]
path = "test"
data_dir = "/app/data" # Don't change if you run podsync via docker

# Tokens from `Access tokens` section
[tokens]
youtube = "YOUTUBE_API_TOKEN" # YouTube API Key. See https://developers.google.com/youtube/registering_an_application
vimeo = [ # Multiple keys will be rotated.
  "VIMEO_API_KEY_1", # Vimeo developer keys. See https://developer.vimeo.com/api/guides/start#generate-access-token
  "VIMEO_API_KEY_2"
]

[feeds]
  [feeds.ID1]
  url = "{FEED_URL}" # URL address of a channel, group, user, or playlist. 
  page_size = 50 # The number of episodes to query each update (keep in mind, that this might drain API token)
  update_period = "12h" # How often query for updates, examples: "60m", "4h", "2h45m"
  quality = "high" # or "low"
  format = "video" # or "audio"
  # custom.cover_art_quality use "high" or "low" to special cover image quality from channel cover default is equal with "quality" and disable when custom.cover_art was set.
  # custom = { title = "Level1News", description = "News sections of Level1Techs, in a podcast feed!", author = "Level1Tech", cover_art = "{IMAGE_URL}", cover_art_quality = "high", category = "TV", subcategories = ["Documentary", "Tech News"], explicit = true, lang = "en" } # Optional feed customizations
  # max_height = 720 # Optional maximal height of video, example: 720, 1080, 1440, 2160, ...
  # cron_schedule = "@every 12h" # Optional cron expression format. If set then overwrite 'update_period'. See details below
  # filters = { title = "regex for title here", not_title = "regex for negative title match", description = "...", not_description = "..." } # Optional Golang regexp format. If set, then only download matching episodes.
  # opml = true|false # Optional inclusion of the feed in the OPML file (default value: false)
  # clean = { keep_last = 10 } # Keep last 10 episodes (order desc by PubDate)
  # youtube_dl_args = [ "--write-sub", "--embed-subs", "--sub-lang", "en,en-US,en-GB" ] # Optional extra arguments passed to youtube-dl when downloading videos from this feed. This example would embed available English closed captions in the videos. Note that setting '--audio-format' for audio format feeds, or '--format' or '--output' for any format may cause unexpected behaviour. You should only use this if you know what you are doing, and have read up on youtube-dl's options!

[database]
  badger = { truncate = true, file_io = true } # See https://github.com/dgraph-io/badger#memory-usage

[downloader]
self_update = true # Optional, auto update youtube-dl every 24 hours
timeout = 15 # Timeout in minutes

# Optional log config. If not specified logs to the stdout
[log]
filename = "podsync.log"
max_size = 50 # MB
max_age = 30 # days
max_backups = 7
compress = true

```

Please note: Automatically clean-up will not work without a database configuration.

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

