FROM alpine:3.10

LABEL org.opencontainers.image.source=https://github.com/fqx/podsync-with-yt-dlp

WORKDIR /app/
RUN wget -O /usr/bin/youtube-dl https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp && \
    chmod +x /usr/bin/youtube-dl && \
    wget -O /app/podsync.tar.gz https://github.com/mxpv/podsync/releases/download/v2.4.0/Podsync_2.4.0_Linux_x86_64.tar.gz && \
    cd /app && tar -xzf podsync.tar.gz && \
    chmod +x /app/podsync && \
    apk --no-cache add ca-certificates python3 ffmpeg tzdata
# Clean up
RUN rm -f /app/podsync.tar.gz

ENTRYPOINT ["/app/podsync"]
CMD ["--no-banner"]
