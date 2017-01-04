$OUTPUT_DIR = 'bin/Publish'

& "dotnet" restore

Remove-Item $OUTPUT_DIR -Force -Recurse -ErrorAction Ignore

& "dotnet" publish --configuration release --output $OUTPUT_DIR

& "docker-compose" up -d --build