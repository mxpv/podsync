# Build fronend
FROM node:11-slim AS builder
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build

# Build nginx
FROM umputun/nginx-le:latest

COPY cmd/nginx/podsync.conf /etc/nginx/service.conf
COPY cmd/nginx/robots.txt /var/www/

COPY --from=builder /app/dist/html/ /var/www/
COPY --from=builder /app/dist/assets/ /var/www/assets/
