# This is a template to be used by GoReleaser.
# See docs for details: https://goreleaser.com/customization/docker/

FROM alpine:3.10
RUN wget -O /usr/bin/youtube-dl https://github.com/ytdl-org/youtube-dl/releases/latest/download/youtube-dl && \
    chmod +x /usr/bin/youtube-dl && \
    apk --no-cache add ca-certificates python ffmpeg tzdata
COPY podsync /podsync

ENTRYPOINT ["/podsync"]
CMD ["--no-banner"]
