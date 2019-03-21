import youtube_dl

BEST_FORMAT = "bestvideo+bestaudio/best"

FORMATS = {
    'video_high': 'best[ext=mp4]',
    'video_low': 'worst[ext=mp4]',
    'audio_high': 'bestaudio',
    'audio_low': 'worstaudio',
}


def handler(event, context):
    url = event['url']
    if not url:
        raise ValueError('Invalid resource URL %s' % url)

    start = int(event['start'])
    if not start or start is None or start < 1:
        start = 1

    end = int(event['end'])
    if end > 600:
        end = 600

    if start > end:
        raise ValueError('Invalid start/end range')

    kind = event['kind']
    if not kind:
        kind = 'video_high'

    last_id = event['last_id']

    print('Getting updated for %s (start=%d, end=%d, kind: %s, last id: %s)', url, start, end, kind, last_id)
    return _get_updates(start, end, url, kind, last_id)


def _get_updates(start, end, url, kind, last_id=None):
    opts = {
        'playliststart': start,
        'playlistend': end,
        'extract_flat': 'in_playlist',
        'quiet': True,
        'no_warnings': True,
        'simulate': True,
        'skip_download': True,
    }

    with youtube_dl.YoutubeDL(opts) as ytdl:
        selector = ytdl.build_format_selector(FORMATS[kind])
        feed_info = ytdl.extract_info(url, download=False)

        # Record basic feed metadata
        feed = {
            'id': feed_info.get('id'),
            'title': feed_info.get('uploader'),
            'page_url': feed_info.get('webpage_url'),
        }

        videos = []
        new_last_id = None

        entries = feed_info['entries']
        for idx, entry in enumerate(entries):
            # Query video metadata from YouTube
            result = ytdl.process_ie_result(entry, download=False)

            video_id = result['id']

            # If already seen this video previously, stop pulling updates
            if last_id and video_id == last_id:
                break

            # Remember new last id
            if idx == 0:
                new_last_id = video_id

            videos.append({
                'id': video_id,
                'title': result.get('title'),
                'description': result.get('description'),
                'thumbnail': result.get('thumbnail'),
                'duration': result.get('duration'),
                'video_url': result.get('webpage_url'),
                'upload_date': result.get('upload_date'),
                'ext': result.get('ext'),
                'size': _get_size(result, selector),
            })

    return {
        'feed': feed,
        'items': videos,
        'last_id': new_last_id,
    }


def _get_size(video, selector):
    try:
        selected = next(selector(video))
    except KeyError:
        selected = video

    if 'requested_formats' in selected:
        return sum(int(f['filesize']) for f in selected['requested_formats'])

    if selected.get('filesize') is not None:
        return int(selected['filesize'])

    return 0
