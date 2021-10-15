FROM alpine:3.10

WORKDIR /app/
RUN wget -O /usr/bin/youtube-dl https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp && \
    chmod +x /usr/bin/youtube-dl && \
    apk --no-cache add ca-certificates python ffmpeg tzdata
COPY podsync /app/podsync

ENTRYPOINT ["/app/podsync"]
CMD ["--no-banner"]
