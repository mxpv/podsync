## Handling static files

`ASSETS_PATH` should point to a directory with static files.
For debugging just 'Copy Path' to `assets` directory.

For production run `gulp patch`.
Gulp will generate `dist` directory with minified files and update templates to include these files.

`TEMPLATES_PATH` should just point to `templates` directory.

Docker will run `gulp` and include `dist` and `templates` directories during build as well as specify `ASSETS_PATH` and `TEMPLATES_PATH` environment variables.

## Patreon

In order to login via Patreon the following variables should be configured:
- `PATREON_REDIRECT_URL` should point to `http://yout_host_here/patreon`
- `PATREON_CLIENT_ID` and `PATREON_SECRET` should be copied from https://www.patreon.com/platform/documentation/clients

## Deploy Docker images

Build docker image:
```
docker build -t ytdl .
```

Deploy image to Container Registry:
```
docker tag ytdl gcr.io/pod-sync/ytdl
gcloud auth application-default login
gcloud docker -- push gcr.io/pod-sync/ytdl
```

or just use `build.sh`