#!/usr/bin/env bash

docker build -t mxpv/nginx -f cmd/nginx/Dockerfile .
docker push mxpv/nginx
