FROM golang:1.11.2 AS backend_builder
WORKDIR /podsync
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go build  -o server -v ./cmd/app

FROM alpine
RUN apk --update --no-cache add ca-certificates
WORKDIR /app/
COPY --from=backend_builder /podsync/server .
ENTRYPOINT ["/app/server"]