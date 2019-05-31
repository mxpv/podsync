import youtube_dl
from datetime import datetime

BEST_FORMAT = "bestvideo+bestaudio/best"
DEFAULT_PAGE_SIZE = 50


def get_format(fmt, quality):
    if fmt == 'video':
        # Video
        if quality == 'high':
            return 'best[ext=mp4]'
        else:
            return 'worst[ext=mp4]'
    else:
        # Audio
        if quality == 'high':
            return 'bestaudio'
        else:
            return 'worstaudio'


def get_updates(start, count, url, fmt, last_id=None, playlist=None):
    if start < 1:
        raise ValueError('Invalid start value')

    if count < 1 or count > 600:
        raise ValueError('Invalid count value')

    if not url:
        raise ValueError('Invalid resource URL %s' % url)

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

    dirty = False

    with youtube_dl.YoutubeDL(opts) as ytdl:
        selector = ytdl.build_format_selector(fmt)
        feed_info = ytdl.extract_info(url, download=False)

        videos = []

        entries = feed_info['entries']
        if not len(entries):  # No episodes
            return videos, None

        is_playlist = len(playlist) > 0
        playlist_lookup = {}
        if is_playlist:
            for item in playlist:
                playlist_lookup[item['ID']] = item

        for idx, entry in enumerate(entries):
            video_id = entry['id']

            if is_playlist:
                if video_id in playlist_lookup:
                    videos.append(playlist_lookup[video_id])
                    continue
            else:
                # If already seen this video previously, stop pulling updates
                if last_id and video_id == last_id:
                    break

            dirty = True

            # Query video metadata from YouTube
            try:
                result = ytdl.process_ie_result(entry, download=False)
            except youtube_dl.utils.DownloadError:
                continue

            # Convert '20190101' to unix time
            date_str = result.get('upload_date')
            date = datetime.strptime(date_str, '%Y%m%d')

            # Duration in seconds
            duration = int(result.get('duration'))
            size = _get_size(result, selector, fmt, duration)

            videos.append({
                'ID': video_id,
                'Title': result.get('title'),
                'Description': result.get('description'),
                'Thumbnail': result.get('thumbnail'),
                'Duration': duration,
                'VideoURL': result.get('webpage_url'),
                'PubDate': int(date.timestamp()),
                'Size': size,
            })

    idx = 0 if not playlist else -1
    new_last_id = entries[idx]['id']
    return videos, new_last_id, dirty


def _get_size(video, selector, fmt, duration):
    try:
        selected = next(selector(video))
    except KeyError:
        selected = video

    if 'requested_formats' in selected:
        return sum(int(f['filesize']) for f in selected['requested_formats'])

    if selected.get('filesize') is not None:
        return int(selected['filesize'])

    # Calculate approximate file size

    is_high = 'best' in fmt
    is_audio = 'audio' in fmt

    if is_audio:
        return (16000 if is_high else 6000) * duration
    else:
        return (350000 if is_high else 100000) * duration
