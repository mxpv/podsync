FROM alpine:3.10
RUN apk --no-cache add \
    ca-certificates \
    youtube-dl \
    ffmpeg
WORKDIR /app/
COPY podsync /app/podsync
ENTRYPOINT ["/app/podsync"]
