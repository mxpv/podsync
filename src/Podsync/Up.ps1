$OUTPUT_DIR = 'bin/Publish'

& "dotnet" restore

Remove-Item $OUTPUT_DIR -Force -Recurse -ErrorAction Ignore

& "dotnet" publish --configuration release --output $OUTPUT_DIR

& "docker" build -t podsync .

# docker run -d -p 5001:5001 podsync

& "docker-compose" up -d