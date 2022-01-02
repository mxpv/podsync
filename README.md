# Podsync

![Podsync](docs/img/logo.png)

[![](https://github.com/mxpv/podsync/workflows/CI/badge.svg)](https://github.com/mxpv/podsync/actions?query=workflow%3ACI)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/mxpv/podsync)](https://github.com/mxpv/podsync/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/mxpv/podsync)](https://goreportcard.com/report/github.com/mxpv/podsync)
[![GitHub Sponsors](https://img.shields.io/github/sponsors/mxpv)](https://github.com/sponsors/mxpv)
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

## Documentation

- [How to get Vimeo API token](./docs/how_to_get_vimeo_token.md)
- [How to get YouTube API Key](./docs/how_to_get_youtube_api_key.md)
- [Podsync on QNAP NAS Guide](./docs/how_to_setup_podsync_on_qnap_nas.md)
- [Schedule updates with cron](./docs/cron.md)

### Access tokens

In order to query YouTube or Vimeo API you have to obtain an API token first.

- [How to get YouTube API key](https://elfsight.com/blog/2016/12/how-to-get-youtube-api-key-tutorial/)
- [Generate an access token for Vimeo](https://developer.vimeo.com/api/guides/start#generate-access-token)

## Configuration

You need to create a configuration file (for instance `config.toml`) and specify the list of feeds that you're going to host.
See [config.toml.example](./config.toml.example) for all possible configuration keys available in Podsync.

Minimal configuration would look like this:

```toml
[server]
port = 8080
data_dir = "/data/podsync/"

[tokens]
youtube = "PASTE YOUR API KEY HERE"

[feeds]
    [feeds.ID1]
    url = "https://www.youtube.com/channel/UCxC5Ls6DwqV0e-CYcAKkExQ"
```

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

