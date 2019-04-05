import youtube_dl
import boto3
import os
import time
from datetime import datetime

BEST_FORMAT = "bestvideo+bestaudio/best"
DEFAULT_PAGE_SIZE = 50

dynamodb = boto3.resource('dynamodb')

feeds_table = dynamodb.Table(os.getenv('UPDATER_DYNAMO_FEEDS_TABLE', 'Feeds'))


def handler(event, context):
    url = event.get('url', None)
    if not url:
        raise ValueError('Invalid resource URL %s' % url)

    start = event.get('start', 1)
    count = event.get('count', DEFAULT_PAGE_SIZE)

    kind = event.get('kind', 'video_high')
    last_id = event.get('last_id', None)

    print('Getting updated for %s (start=%d, count=%d, kind: %s, last id: %s)' % (url, start, count, kind, last_id))
    return _get_updates(start, count, url, kind, last_id)


def _update_feed(hash_id):
    print('Updating feed {}'.format(hash_id))
    feed = _query_feed(hash_id)

    page_size = int(feed.get('PageSize', DEFAULT_PAGE_SIZE))
    last_id = feed.get('LastID', None)
    episodes = feed.get('Episodes', [])

    # Rebuild episode list from scratch
    if not last_id:
        episodes = []

    start = time.time()
    _, items, new_last_id = _get_updates(1, page_size, _get_url(feed), _get_format(feed), last_id)
    end = time.time()

    print('Got feed update: new {}, current {}. Update took: {}'.format(len(items), len(episodes), end-start))

    # Update feed and submit back to Dynamo

    unix_time = int(datetime.utcnow().timestamp())
    feed['UpdatedAt'] = unix_time

    if len(items) > 0:
        episodes = items + episodes  # Prepand new episodes
        del episodes[page_size:]  # Truncate list
        feed['Episodes'] = episodes

        # Update last seen video ID
        feed['LastID'] = new_last_id

        _update_feed_episodes(hash_id, feed)
    else:
        # Update last access field only
        _update_feed_updated_at(hash_id, unix_time)


def _query_feed(hash_id):
    response = feeds_table.get_item(
        Key={'HashID': hash_id},
        ProjectionExpression='#prov,#type,#size,#fmt,#quality,#level,#id,#last_id,#episodes,#updated_at',
        ExpressionAttributeNames={
            '#prov': 'Provider',
            '#type': 'LinkType',
            '#size': 'PageSize',
            '#fmt': 'Format',
            '#quality': 'Quality',
            '#level': 'FeatureLevel',
            '#id': 'ItemID',
            '#last_id': 'LastID',
            '#episodes': 'Episodes',
            '#updated_at': 'UpdatedAt',
        },
    )

    item = response['Item']
    return item


def _update_feed_episodes(hash_id, feed):
    feeds_table.update_item(
        Key={
            'HashID': hash_id,
        },
        UpdateExpression='SET #updated_at = :updated_at, #episodes = :episodes, #last_id = :last_id',
        ExpressionAttributeNames={
            '#updated_at': 'UpdatedAt',
            '#episodes': 'Episodes',
            '#last_id': 'LastID',
        },
        ExpressionAttributeValues={
            ':updated_at': feed['UpdatedAt'],
            ':episodes': feed['Episodes'],
            ':last_id': feed['LastID'],
        },
        ReturnValues='NONE',
    )


def _update_feed_updated_at(hash_id, updated_at):
    feeds_table.update_item(
        Key={
            'HashID': hash_id,
        },
        UpdateExpression='SET #updated_at = :updated_at',
        ExpressionAttributeNames={
            '#updated_at': 'UpdatedAt',
        },
        ExpressionAttributeValues={
            ':updated_at': updated_at,
        },
        ReturnValues='NONE',
    )


def _get_format(feed):
    fmt = feed.get('Format', 'video')
    quality = feed.get('Quality', 'high')

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


def _get_url(feed):
    provider = feed['Provider']
    link_type = feed['LinkType']
    item_id = feed['ItemID']

    if provider == 'youtube':

        if link_type == 'playlist':
            return 'https://www.youtube.com/playlist?list={}'.format(item_id)
        elif link_type == 'channel':
            return 'https://www.youtube.com/channel/{}'.format(item_id)
        elif link_type == 'user':
            return 'https://www.youtube.com/user/{}'.format(item_id)
        else:
            raise ValueError('Unsupported link type')

    elif provider == 'vimeo':

        if link_type == 'channel':
            return 'https://vimeo.com/channels/{}'.format(item_id)
        elif link_type == 'group':
            return 'http://vimeo.com/groups/{}'.format(item_id)
        elif link_type == 'user':
            return 'https://vimeo.com/{}'.format(item_id)
        else:
            raise ValueError('Unsupported link type')

    else:
        raise ValueError('Unsupported provider')


def _get_updates(start, count, url, fmt, last_id=None):
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
        selector = ytdl.build_format_selector(fmt)
        feed_info = ytdl.extract_info(url, download=False)

        # Record basic feed metadata
        feed = {
            'ID': feed_info.get('id'),
            'Title': feed_info.get('uploader'),
            'PageURL': feed_info.get('webpage_url'),
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
                'ID': video_id,
                'Title': result.get('title'),
                'Description': result.get('description'),
                'Thumbnail': result.get('thumbnail'),
                'Duration': result.get('duration'),
                'VideoURL': result.get('webpage_url'),
                'UploadDate': result.get('upload_date'),
                'Ext': result.get('ext'),
                'Size': _get_size(result, selector),
            })

    return feed, videos, new_last_id


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
