FROM golang:1.17.5-alpine3.15 AS builder
WORKDIR /app
COPY . /app/
RUN go build -o podsync ./cmd/podsync

FROM alpine:3.10
WORKDIR /app/
RUN wget -O /usr/bin/youtube-dl https://github.com/ytdl-org/youtube-dl/releases/latest/download/youtube-dl && \
    chmod +x /usr/bin/youtube-dl && \
    apk --no-cache add ca-certificates python ffmpeg tzdata
COPY --from=builder /app/podsync /app/podsync

ENTRYPOINT ["/app/podsync"]
CMD ["--no-banner"]
