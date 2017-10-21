FROM python:3.6.0-alpine

WORKDIR /app
COPY . .

ENV REDIS_CONNECTION_STRING localhost

RUN apk add --update --no-cache gcc g++ make && \
    pip install --no-cache-dir --requirement /app/requirements.txt

EXPOSE 5002
ENTRYPOINT ["python", "/app/ytdl.py"]