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

opts = {
    'quiet': True,
    'no_warnings': True,
    'forceurl': True,
    'simulate': True,
    'skip_download': True,
    'call_home': False,
    'nocheckcertificate': True
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

    try:
        with youtube_dl.YoutubeDL(opts) as ytdl:
            info = ytdl.extract_info(url, download=False)
            return _choose_url(info, quality)
    except DownloadError:
        raise
    except Exception as e:
        print(e)
        raise


def _choose_url(info, quality):
    is_video = quality == 'videohigh' or quality == 'videolow'

    # Filter formats by file extension
    ext = 'mp4' if is_video else 'm4a'
    fmt_list = [x for x in info['formats'] if x['ext'] == ext and x['acodec'] != 'none']
    if not len(fmt_list):
        return info['url']

    # Sort list by field (width for videos, file size for audio)
    sort_field = 'width' if is_video else 'filesize'
    ordered = sorted(fmt_list, key=lambda x: x[sort_field] or x['format_id'], reverse=True)

    # Choose an item depending on quality, better at the beginning
    is_high_quality = quality == 'videohigh' or quality == 'audiohigh'
    item = ordered[0] if is_high_quality else ordered[-1]
    return item['url']


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5002, workers=32)
