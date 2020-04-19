FROM alpine:3.10

WORKDIR /app/
RUN \
    apk --no-cache add ca-certificates python py-pip ffmpeg && \
    pip install youtube-dl
COPY podsync /app/podsync
CMD ["/app/podsync"]
