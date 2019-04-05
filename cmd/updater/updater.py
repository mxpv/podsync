import youtube_dl

BEST_FORMAT = "bestvideo+bestaudio/best"

FORMATS = {
    'video_high': 'best[ext=mp4]',
    'video_low': 'worst[ext=mp4]',
    'audio_high': 'bestaudio',
    'audio_low': 'worstaudio',
}


def handler(event, context):
    url = event.get('url', None)
    if not url:
        raise ValueError('Invalid resource URL %s' % url)

    start = event.get('start', 1)
    count = event.get('count', 50)

    kind = event.get('kind', 'video_high')
    last_id = event.get('last_id', None)

    print('Getting updated for %s (start=%d, count=%d, kind: %s, last id: %s)' % (url, start, count, kind, last_id))
    return _get_updates(start, count, url, kind, last_id)


def _get_updates(start, count, url, kind, last_id=None):
    if start < 1:
        raise ValueError('Invalid start value')

    if count < 1 or count > 600:
        raise ValueError('Invalid count value')

    end = start + count - 1

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
            video_id = entry['id']

            # If already seen this video previously, stop pulling updates
            if last_id and video_id == last_id:
                break

            # Remember new last id
            if idx == 0:
                new_last_id = video_id

            # Query video metadata from YouTube
            result = ytdl.process_ie_result(entry, download=False)

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
