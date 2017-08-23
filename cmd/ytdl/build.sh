docker build -t ytdl .
docker tag ytdl gcr.io/pod-sync/ytdl
gcloud docker -- push gcr.io/pod-sync/ytdl