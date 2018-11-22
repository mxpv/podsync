FROM node:latest AS gulp
WORKDIR /app
COPY . .
RUN npm install
RUN npm link gulp
RUN gulp patch

FROM golang:1.11.2 AS build
WORKDIR /podsync
COPY --from=gulp /app .
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go build  -o server -v ./cmd/app

FROM alpine
RUN apk --update --no-cache add ca-certificates
WORKDIR /app/
COPY --from=gulp /app/templates ./templates
COPY --from=gulp /app/dist ./assets
COPY --from=build /podsync/server .
ENV ASSETS_PATH /app/assets
ENV TEMPLATES_PATH /app/templates
ENTRYPOINT ["/app/server"]