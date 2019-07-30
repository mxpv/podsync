from resolver import *


def handler(event, lambda_context):
    feed_id = event['feed_id']
    video_id = event['video_id']

    redirect_url = download(feed_id, video_id)
    return {
        'code': 302,
        'redirect_url': redirect_url,
    }
