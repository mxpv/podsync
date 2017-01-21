#!/bin/bash

OUTPUT_DIR=${1:-'bin/Publish'}

rm -rf $OUTPUT_DIR

dotnet restore
dotnet publish --configuration release --output $OUTPUT_DIR

docker-compose up -d --build