# This is a template to be used by GoReleaser.
# See docs for details: https://goreleaser.com/customization/docker/

FROM alpine:3.16
WORKDIR /app

RUN wget -O /usr/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp && \
    chmod +x /usr/bin/yt-dlp && \
    ln -s /usr/bin/yt-dlp /usr/bin/youtube-dl && \
    apk --no-cache add ca-certificates python3 py3-pip ffmpeg tzdata
COPY podsync /app/podsync

ENTRYPOINT ["/app/podsync"]
CMD ["--no-banner"]
