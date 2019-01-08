![Build Status](https://codebuild.us-east-1.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoiOE52ek1HRWdQeW5pVmozMUNtWm1zcXBDc0FPbTFRRzZRWEhQeGFrOXd6TFFhVnlVOHQ0dWM5SHFZRnloQUFKOUY2NWdMaDBOdnMxUnYyYW9FZC9GbElNPSIsIml2UGFyYW1ldGVyU3BlYyI6ImVwQkN3WjV4MXpTZ2FXTUUiLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=master)
[![Patreon](https://img.shields.io/badge/support-patreon-E6461A.svg)](https://www.patreon.com/podsync)


## Patreon

In order to login via Patreon the following variables should be configured:
- `PATREON_REDIRECT_URL` should point to `http://yout_host_here/patreon`
- `PATREON_CLIENT_ID` and `PATREON_SECRET` should be copied from https://www.patreon.com/platform/documentation/clients

## Building Docker images

Backend
```bash
./build.sh
```

nginx
```bash
./nginx.sh
```

ytdl
```bash
cd cmd/ytdl/
./build.sh
```

## Running
```bash
docker-compose pull
docker-compose up -d
```
