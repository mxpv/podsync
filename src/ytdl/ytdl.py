import youtube_dl
import redis
import os
from youtube_dl.utils import DownloadError
from sanic import Sanic
from sanic.exceptions import InvalidUsage, NotFound
from sanic.response import text, redirect
from datetime import timedelta

app = Sanic()

db = redis.from_url(os.getenv('REDIS_CONNECTION_STRING', 'redis://localhost:6379'))
db.ping()

default_opts = {
    'quiet': True,
    'no_warnings': True,
    'forceurl': True,
    'simulate': True,
    'skip_download': True,
    'call_home': False,
    'nocheckcertificate': True
}

youtube_quality = {
    'videohigh': 'best[ext=mp4]',
    'videolow': 'worst[ext=mp4]',
    'audiohigh': 'bestaudio[ext=m4a]/worstaudio[ext=m4a]',
    'audiolow': 'worstaudio[ext=m4a]/bestaudio[ext=m4a]'
}

vimeo_quality = {
    'videohigh': 'Original/http-1080p/http-720p/http-360p/http-270p',
    'videolow': 'http-270p/http-360p/http-540p/http-720p/http-1080p'
}

url_formats = {
    'youtube': 'https://youtube.com/watch?v={}',
    'vimeo': 'https://vimeo.com/{}',
}


@app.route('/download/<feed_id>/<video_id>', methods=['GET'])
async def download(request, feed_id, video_id):
    if not feed_id:
        raise InvalidUsage('Invalid feed id')

    # Remote extension and check if video id is ok
    video_id = os.path.splitext(video_id)[0]
    if not video_id:
        raise InvalidUsage('Invalid video id')

    # Query redis
    data = db.hgetall(feed_id)
    if not data:
        raise NotFound('Feed not found')

    # Delete this feed if no requests within 90 days
    db.expire(feed_id, timedelta(days=90))

    entries = {k.decode().lower(): v.decode().lower() for k, v in data.items()}

    # Build URL
    provider = entries.get('provider')
    tpl = url_formats[provider]
    if not tpl:
        raise InvalidUsage('Invalid feed')

    url = tpl.format(video_id)
    quality = entries.get('quality')

    try:
        redirect_url = _resolve(url, quality)
        return redirect(redirect_url)
    except DownloadError as e:
        msg = str(e)
        return text(msg, status=511)


def _resolve(url, quality):
    if not url:
        raise InvalidUsage('Invalid URL')

    if not quality:
        quality = 'videohigh'

    opts = default_opts.copy()
    fmt = _choose_format(quality, url)

    if fmt:
        opts.update(format=fmt)

    try:
        with youtube_dl.YoutubeDL(opts) as ytdl:
            info = ytdl.extract_info(url, download=False)
            return info['url']
    except DownloadError:
        raise
    except Exception as e:
        print(e)
        raise


def _choose_format(quality, url):
    fmt = None
    if 'youtube.com' in url:
        fmt = youtube_quality.get(quality)
    elif 'vimeo.com' in url:
        fmt = vimeo_quality.get(quality)

    return fmt


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5002)
