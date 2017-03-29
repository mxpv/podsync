FROM microsoft/aspnetcore:1.1.1

ARG PUBLISH_DIR="bin/Publish"

WORKDIR /app
COPY $PUBLISH_DIR .

ENTRYPOINT ["dotnet", "Podsync.dll"]