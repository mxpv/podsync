#!/bin/bash

OUTPUT_DIR=${1:-'bin/Publish'}

rm -rf src/Podsync/$OUTPUT_DIR

dotnet restore
dotnet publish --configuration release --output $OUTPUT_DIR

docker-compose up -d --build