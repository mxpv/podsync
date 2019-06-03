# Build fronend
FROM node:11-slim AS builder
WORKDIR /work
COPY ui/ .
RUN npm install
RUN npm run build -- --prod --aot

# Build nginx to host static files
FROM nginx:stable

COPY cmd/nginx/podsync.conf /etc/nginx/nginx.conf

COPY cmd/nginx/robots.txt /var/www/podsync/
COPY --from=builder /work/dist/ui/ /var/www/podsync/
