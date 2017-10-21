import os
import requests
import youtube_dl
from sanic import Sanic
from sanic.exceptions import InvalidUsage
from sanic.response import text, redirect
from youtube_dl.utils import DownloadError

METADATA_URL = os.getenv('METADATA_URL', 'http://app:5001/api/metadata/{feed_id}')
print('Using metadata URL template: ' + METADATA_URL)

app = Sanic()

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


@app.route('/download/<feed_id>/<video_id>', methods=['GET', 'HEAD'])
async def download(req, feed_id, video_id):
    if not feed_id:
        raise InvalidUsage('Invalid feed id')

    # Remote extension and check if video id is ok
    video_id = os.path.splitext(video_id)[0]
    if not video_id:
        raise InvalidUsage('Invalid video id')

    # Pull metadata from API server
    metadata_url = METADATA_URL.format(feed_id=feed_id, video_id=video_id)
    r = requests.get(url=metadata_url)
    json = r.json()

    # Build URL
    provider = json['provider']
    tpl = url_formats[provider]
    if not tpl:
        raise InvalidUsage('Invalid feed')
    url = tpl.format(video_id)

    try:
        redirect_url = _resolve(url, json)
        return redirect(redirect_url)
    except DownloadError as e:
        msg = str(e)
        return text(msg, status=511)


def _resolve(url, metadata):
    if not url:
        raise InvalidUsage('Invalid URL')

    try:
        provider = metadata['provider']

        with youtube_dl.YoutubeDL(opts) as ytdl:
            info = ytdl.extract_info(url, download=False)
            if provider == 'youtube':
                return _choose_url(info, metadata)
            elif provider == 'vimeo':
                return info['url']
            else:
                raise ValueError('undefined provider')
    except DownloadError:
        raise
    except Exception as e:
        print(e)
        raise


def _choose_url(info, metadata):
    is_video = metadata['format'] == 'video'

    # Filter formats by file extension
    ext = 'mp4' if is_video else 'm4a'
    fmt_list = [x for x in info['formats'] if x['ext'] == ext and 'acodec' in x and x['acodec'] != 'none']
    if not len(fmt_list):
        return info['url']

    # Sort list by field (width for videos, file size for audio)
    sort_field = 'width' if is_video else 'filesize'
    # Sometime 'filesize' field can be None
    if not all(x[sort_field] is not None for x in fmt_list):
        sort_field = 'format_id'
    ordered = sorted(fmt_list, key=lambda x: x[sort_field], reverse=True)

    # Choose an item depending on quality, better at the beginning
    is_high_quality = metadata['quality'] == 'high'
    item = ordered[0] if is_high_quality else ordered[-1]
    return item['url']


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5002, workers=32)