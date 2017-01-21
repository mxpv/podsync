FROM microsoft/aspnetcore:1.0.3

ARG PUBLISH_DIR="bin/Publish"

WORKDIR /app

COPY $PUBLISH_DIR .

RUN apt-get update                          && \
    apt-get install -y python-setuptools    && \
    easy_install pip                        && \
    pip install youtube-dl

ENTRYPOINT ["dotnet", "Podsync.dll"]