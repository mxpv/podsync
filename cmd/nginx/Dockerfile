# Build fronend
FROM node:8-slim AS frontend_builder
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build

# Build nginx
FROM umputun/nginx-le:latest

COPY cmd/nginx/podsync.conf /etc/nginx/service.conf
COPY cmd/nginx/robots.txt /var/www/

COPY --from=frontend_builder /app/templates/ /var/www/
COPY --from=frontend_builder /app/dist/ /var/www/assets/
