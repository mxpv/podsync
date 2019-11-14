# Podsync

![Podsync](docs/img/logo.png)

[![Build Status](https://travis-ci.com/mxpv/podsync.svg?branch=master)](https://travis-ci.com/mxpv/podsync)
[![Go Report Card](https://goreportcard.com/badge/github.com/mxpv/podsync)](https://goreportcard.com/report/github.com/mxpv/podsync)
[![Patreon](https://img.shields.io/badge/support-patreon-E6461A.svg)](https://www.patreon.com/podsync)

Podsync - is a simple, free service that lets you listen to any YouTube / Vimeo channels, playlists or user videos in
podcast format.

Podcast applications have a rich functionality for content delivery - automatic download of new episodes,
remembering last played position, sync between devices and offline listening. This functionality is not available
on YouTube and Vimeo. So the aim of Podsync is to make your life easier and enable you to view/listen to content on
any device in podcast client.

## Dependencies

If you're running the CLI as binary (e.g. not via Docker), you need to make sure that dependencies are available on
your system. Currently Podsync depends on `youtube-dl` and `ffmpeg`.

On Mac you can install those with `brew`:
```
brew install youtube-dl ffmpeg
```

## Access tokens

In order to query YouTube or Vimeo API you have to obtain an API token first.

- [How to get YouTube API key](https://elfsight.com/help/how-to-get-youtube-api-key/)
- [Generate an access token for Vimeo](https://developer.vimeo.com/api/guides/start#generate-access-token)

## Configuration example

```toml
[server]
port = 8080
data_dir = "/path/to/data/directory"
hostname = "http://hostname"

[tokens]
youtube = "{YOUTUBE_API_TOKEN}"
vimeo = "{VIMEO_API_TOKEN}"

[feeds]
  [feeds.ID1]
  url = "{FEED_URL}" # URL address of a channel, group, user, or playlist. 
  page_size = 50 # The number of episodes to query each update (keep in mind, that this might drain API token)
  update_period = "12h" # How often query for updates, examples: "60m", "4h", "2h45m"
  quality = "high" # or "low"
  format = "video" # or "low"
```

Episodes files will be kept at: `/path/to/data/directory/ID1`

Feed will be accessible from: `http://hostname/ID1.xml`

## How to run

Run as binary:
```
$ ./podsync --config config.toml
```

Run via Docker:
```
$ docker pull mxpv/podsync:latest
$ docker run \
    -p 8080:8080 \
    -v $(pwd)/data:/app/data/ \
    -v $(pwd)/config.toml:/app/config.toml \
    mxpv/podsync:latest
```

Run via Docker Compose:
```
$ docker-compose up
```

## How to make a release

- Add and push version tag
  ```
  $ git tag -a v0.1.0 -m "First release"
  $ git push origin --tags
  ```
- Run GoReleaser at the root of your repository:
  ```
  $ goreleaser --rm-dist
  ```
