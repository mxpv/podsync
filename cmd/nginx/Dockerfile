# Build fronend
FROM node:11-slim AS builder
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build

# Build nginx to host static files
FROM nginx:stable

COPY cmd/nginx/podsync.conf /etc/nginx/nginx.conf

COPY cmd/nginx/robots.txt /var/www/podsync/
COPY --from=builder /app/dist/html/ /var/www/podsync/
COPY --from=builder /app/dist/assets/ /var/www/podsync/assets/
