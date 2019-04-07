from updater import DEFAULT_PAGE_SIZE, _get_updates, _get_format


# AWS Lambda entry point
def handler(event, context):
    url = event.get('url', None)
    start = event.get('start', 1)
    count = event.get('count', DEFAULT_PAGE_SIZE)

    # Last seen video ID
    last_id = event.get('last_id', None)

    # Detect item format
    fmt = event.get('format', 'video')
    quality = event.get('quality', 'high')
    ytdl_fmt = _get_format(fmt, quality)

    print('Getting updates for %s (start=%d, count=%d, fmt: %s, last id: %s)' % (url, start, count, ytdl_fmt, last_id))
    _, episodes, new_last_id = _get_updates(start, count, url, ytdl_fmt, last_id)

    return {
        'LastID': new_last_id,
        'Episodes': episodes,
    }
