import os
import requests
import youtube_dl

METADATA_URL = os.getenv('METADATA_URL', 'http://podsync.net/api/metadata/{feed_id}')
print('Using metadata URL template: ' + METADATA_URL)


class InvalidUsage(Exception):
    pass


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


def handler(event, context):
    feed_id = event['feed_id']
    video_id = event['video_id']

    redirect_url = download(feed_id, video_id)

    return {
        'redirect_url': redirect_url,
    }


def download(feed_id, video_id):
    if not feed_id:
        raise InvalidUsage('Invalid feed id')

    # Remove extension and check if video id is ok
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

    redirect_url = _resolve(url, json)
    return redirect_url


def _resolve(url, metadata):
    if not url:
        raise InvalidUsage('Invalid URL')

    print('Resolving %s' % url)

    try:
        provider = metadata['provider']

        with youtube_dl.YoutubeDL(opts) as ytdl:
            info = ytdl.extract_info(url, download=False)
            if provider == 'youtube':
                return _yt_choose_url(info, metadata)
            elif provider == 'vimeo':
                return _vimeo_choose_url(info, metadata)
            else:
                raise ValueError('undefined provider')
    except Exception as e:
        print(e)
        raise


def _yt_choose_url(info, metadata):
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


def _vimeo_choose_url(info, metadata):
    # Query formats with 'extension' = mp4 and 'format_id' = http-1080p/http-720p/../http-360p
    fmt_list = [x for x in info['formats'] if x['ext'] == 'mp4' and x['format_id'].startswith('http-')]

    ordered = sorted(fmt_list, key=lambda x: x['width'], reverse=True)
    is_high_quality = metadata['quality'] == 'high'
    item = ordered[0] if is_high_quality else ordered[-1]

    return item['url']
