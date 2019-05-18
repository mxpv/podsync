FROM python:alpine3.7

WORKDIR /app
COPY . .

RUN apk add --virtual deps --no-cache build-base && \
    pip3 install --no-cache-dir --requirement requirements.txt --target /app && \
    apk del deps && \
    addgroup -S app && adduser -S app -G app

USER app

ENTRYPOINT ["python3", "-m", "sanic", "server.app", "--host", "0.0.0.0", "--port", "8080"]
CMD ["--workers", "1"]
