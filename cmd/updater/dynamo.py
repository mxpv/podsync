import boto3
import os
import time
from updater import DEFAULT_PAGE_SIZE, _get_updates, _get_format
from datetime import datetime

dynamodb = boto3.resource('dynamodb')

feeds_table = dynamodb.Table(os.getenv('UPDATER_DYNAMO_FEEDS_TABLE', 'Feeds'))


def _update_feed(hash_id):
    print('Updating feed {}'.format(hash_id))
    feed = _query_feed(hash_id)

    page_size = int(feed.get('PageSize', DEFAULT_PAGE_SIZE))
    last_id = feed.get('LastID', None)
    episodes = feed.get('Episodes', [])
    item_url = feed['ItemURL']

    # Format parameters
    fmt = feed.get('Format', 'video')
    quality = feed.get('Quality', 'high')

    # Rebuild episode list from scratch
    if not last_id:
        episodes = []

    start = time.time()
    _, items, new_last_id = _get_updates(1, page_size, item_url, _get_format(fmt, quality), last_id)
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
        ProjectionExpression='#prov,#type,#size,#fmt,#quality,#level,#id,#last_id,#episodes,#updated_at,#item_url',
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
            '#item_url': 'ItemURL',
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
