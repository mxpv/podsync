#!/usr/bin/env bash

docker build -t ytdl .
docker tag ytdl mxpv/ytdl
docker push mxpv/ytdl