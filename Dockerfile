FROM alpine:3.10

WORKDIR /app/

RUN apk --no-cache add ca-certificates python ffmpeg tzdata
# see #191 for youtube-dl related questions
RUN apk --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main add youtube-dl 

COPY podsync /app/podsync
CMD ["/app/podsync"]
