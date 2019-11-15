FROM alpine:3.10

WORKDIR /app/
RUN wget -O youtube-dl https://github.com/ytdl-org/youtube-dl/releases/download/2019.11.05/youtube-dl && \
    chmod +x youtube-dl && \
    apk --no-cache add ca-certificates python ffmpeg
COPY podsync /app/podsync
CMD ["/app/podsync"]
