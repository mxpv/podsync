FROM golang:1.12 as build
WORKDIR /work
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 make build

FROM alpine:3.10
RUN apk --no-cache add \
    ca-certificates \
    youtube-dl \
    ffmpeg
WORKDIR /app/
COPY --from=build /work/podsync /app/podsync
ENTRYPOINT ["/app/podsync"]
