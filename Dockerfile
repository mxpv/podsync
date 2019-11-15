FROM alpine:3.10

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk update && \
    apk --no-cache add \
        ca-certificates \
        youtube-dl==2019.11.05-r1 \
        ffmpeg
WORKDIR /app/
COPY podsync /app/podsync
ENTRYPOINT ["/app/podsync"]
