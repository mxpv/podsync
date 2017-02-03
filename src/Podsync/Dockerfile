FROM microsoft/aspnetcore:1.1.0-runtime

ARG PUBLISH_DIR="bin/Publish"

WORKDIR /app
COPY $PUBLISH_DIR .

ENTRYPOINT ["dotnet", "Podsync.dll"]