FROM python:3.6.0-alpine

WORKDIR /app
COPY . .

RUN apk add --update --no-cache gcc g++ make && \
    pip install --no-cache-dir --requirement /app/requirements.txt

EXPOSE 8080
ENTRYPOINT ["python", "/app/ytdl.py"]