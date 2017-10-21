docker build -t app .
docker tag app gcr.io/pod-sync/app
gcloud docker -- push gcr.io/pod-sync/app