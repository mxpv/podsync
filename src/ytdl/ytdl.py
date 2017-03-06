import youtube_dl
from youtube_dl.utils import DownloadError
from sanic import Sanic
from sanic.exceptions import InvalidUsage
from sanic.response import text

app = Sanic()

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
    'VideoHigh': 'best[ext=mp4]',
    'VideoLow': 'worst[ext=mp4]',
    'AudioHigh': 'bestaudio[ext=m4a]/worstaudio[ext=m4a]',
    'AudioLow': 'worstaudio[ext=m4a]/bestaudio[ext=m4a]'
}

vimeo_quality = {
    'VideoHigh': 'Original/http-1080p/http-720p/http-360p/http-270p',
    'VideoLow': 'http-270p/http-360p/http-540p/http-720p/http-1080p'
}


@app.route('/resolve')
async def youtube(request):
    url = request.args.get('url')
    if not url:
        raise InvalidUsage('Invalid URL')

    opts = default_opts.copy()

    quality = request.args.get('quality', 'VideoHigh')
    fmt = _choose_format(quality, url)

    if fmt:
        opts.update(format=fmt)

    try:
        with youtube_dl.YoutubeDL(opts) as ytdl:
            info = ytdl.extract_info(url, download=False)
            return text(info['url'])
    except DownloadError as e:
        msg = str(e)
        return text(msg, status=511)


def _choose_format(quality, url):
    fmt = None
    if 'youtube.com' in url:
        fmt = youtube_quality.get(quality)
    elif 'vimeo.com' in url:
        fmt = vimeo_quality.get(quality)

    return fmt


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)
