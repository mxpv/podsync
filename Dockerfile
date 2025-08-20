FROM golang:1.24 AS builder

ENV TAG="nightly"
ENV COMMIT=""

WORKDIR /build

COPY . .

RUN make build

# Download yt-dlp
RUN wget -O /usr/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp && \
    chmod a+rwx /usr/bin/yt-dlp

# Alpine 3.21 will go EOL on 2026-11-01
FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates python3 py3-pip ffmpeg tzdata libc6-compat

COPY --from=builder /usr/bin/yt-dlp /usr/local/bin/youtube-dl
COPY --from=builder /build/bin/podsync /app/podsync
COPY --from=builder /build/html/index.html /app/html/index.html

ENTRYPOINT ["/app/podsync"]
CMD ["--no-banner"]
