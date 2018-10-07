#!/usr/bin/env bash

docker build -t mxpv/podsync .
docker push mxpv/podsync
