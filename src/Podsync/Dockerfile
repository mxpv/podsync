FROM microsoft/aspnetcore:1.0.3

ARG PORT=56247
ARG PUBLISH_DIR="bin/Publish"

WORKDIR /app

ENV ASPNETCORE_URLS http://+:$PORT
EXPOSE $PORT

COPY $PUBLISH_DIR .

RUN apt-get update                          && \
    apt-get install -y python-setuptools    && \
    easy_install pip                        && \
    pip install youtube-dl

ENTRYPOINT ["dotnet", "Podsync.dll"]