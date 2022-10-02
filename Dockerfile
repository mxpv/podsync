# This is a template to be used by GoReleaser.
# See docs for details: https://goreleaser.com/customization/docker/

FROM golang:alpine

RUN wget -O /usr/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp
RUN chmod +x /usr/bin/yt-dlp
RUN ln -s /usr/bin/yt-dlp /usr/bin/youtube-dl
RUN apk --no-cache add ca-certificates python3 py3-pip ffmpeg tzdata

COPY . /tmp/podsync

WORKDIR /tmp/podsync
RUN go build -o /app/podsync ./cmd/podsync

WORKDIR /app
ENTRYPOINT ["/app/podsync", "--no-banner"]
#CMD ["--no-banner"]
