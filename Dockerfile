FROM alpine:3.10

WORKDIR /app/
RUN wget -O /usr/bin/youtube-dl https://github.com/ytdl-org/youtube-dl/releases/download/2020.01.24/youtube-dl && \
    chmod +x /usr/bin/youtube-dl && \
    apk --no-cache add ca-certificates python ffmpeg
COPY podsync /app/podsync
CMD ["/app/podsync"]
