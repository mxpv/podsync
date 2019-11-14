# Podsync

![Podsync](docs/img/logo.png)

[![Build Status](https://travis-ci.com/mxpv/podsync.svg?branch=master)](https://travis-ci.com/mxpv/podsync)
[![Go Report Card](https://goreportcard.com/badge/github.com/mxpv/podsync)](https://goreportcard.com/report/github.com/mxpv/podsync)
[![Patreon](https://img.shields.io/badge/support-patreon-E6461A.svg)](https://www.patreon.com/podsync)

Podsync - is a simple, free service that lets you listen to any YouTube / Vimeo channels, playlists or user videos in podcast format.

Podcast applications have a rich functionality for content delivery - automatic download of new episodes, remembering last played position, sync between devices and offline listening. This functionality is not available on YouTube and Vimeo. So the aim of Podsync is to make your life easier and enable you to view/listen to content on any device in podcast client.

## How to release

- Add and push version tag
  ```
  $ git tag -a v0.1.0 -m "First release"
  $ git push origin v0.1.0
  ```
- Run GoReleaser at the root of your repository:
  ```
  $ goreleaser --rm-dist
  ```
