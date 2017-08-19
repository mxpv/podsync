FROM golang:1.8 AS build
WORKDIR /go/src/github.com/mxpv/podsync
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go install -v ./cmd/app

FROM alpine
RUN apk --update --no-cache add ca-certificates
WORKDIR /app/
COPY --from=build /go/bin/app .
ENTRYPOINT ["/app/app"]