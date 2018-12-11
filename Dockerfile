FROM node:8-slim AS frontend_builder
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build

FROM golang:1.11.2 AS backend_builder
WORKDIR /podsync
COPY --from=frontend_builder /app .
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go build  -o server -v ./cmd/app

FROM alpine
RUN apk --update --no-cache add ca-certificates
WORKDIR /app/
COPY --from=frontend_builder /app/templates ./templates
COPY --from=frontend_builder /app/dist ./assets
COPY --from=backend_builder /podsync/server .
ENV ASSETS_PATH /app/assets
ENV TEMPLATES_PATH /app/templates
ENTRYPOINT ["/app/server"]