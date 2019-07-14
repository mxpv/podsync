from resolver import *


def handler(event, lambda_context):
    try:
        feed_id, video_id = _get_ids(event.get('path'))
        redirect_url = download(feed_id, video_id)
        return {
            'statusCode': 302,
            'statusDescription': '302 Found',
            'headers': {
                'Location': redirect_url,
                'Content-Type': 'text/plain',
            }
        }
    except QuotaExceeded:
        return {
            'statusCode': 429,
            'statusDescription': '429 Too Many Requests. Daily limit is 100. '
                                 'Consider upgrading account to get unlimited access.',
            'headers': {'Content-Type': 'text/plain'}
        }


def _get_ids(path):
    if not path or not path.startswith('/download'):
        raise InvalidUsage('Invalid path')

    sections = path.split('/')

    # >>> '/download/feed/video.xml'.split('/', 3)
    # ['', 'download', 'feed', 'video.xml']
    if len(sections) != 4:
        raise InvalidUsage('Invalid path')

    feed_id = sections[2]
    video_id = sections[3]

    if not feed_id or not video_id:
        raise InvalidUsage('Invalid feed or video id')

    # Trim extension
    # >>> os.path.splitext('video.xml')[0]
    # 'video'
    video_id = os.path.splitext(video_id)[0]

    return feed_id, video_id
