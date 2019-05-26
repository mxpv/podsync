FROM python:alpine3.7

WORKDIR /app
COPY . .

RUN pip3 install --no-cache-dir --requirement requirements.txt --target /app && \
    addgroup -g 1000 -S app && adduser -u 1000 -G app -S app

USER app

ENTRYPOINT ["python3"]
CMD ["-u", "main.py"]
