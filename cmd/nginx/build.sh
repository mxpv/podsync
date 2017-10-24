#!/usr/bin/env bash

docker build -t nginx .
docker tag nginx gcr.io/pod-sync/nginx
gcloud docker -- push gcr.io/pod-sync/nginx