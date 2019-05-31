import datetime
import gzip
import json
import os
import boto3
import youtube_dl
import updater

sqs = boto3.client('sqs')
sqs_url = os.getenv('UPDATER_SQS_QUEUE_URL')
print('Using SQS URL: {}'.format(sqs_url))

dynamodb = boto3.resource('dynamodb')
feeds_table_name = os.getenv('DYNAMO_FEEDS_TABLE_NAME', 'Feeds')
print('Using DynamoDB table: {}'.format(feeds_table_name))
feeds_table = dynamodb.Table(feeds_table_name)


def _get_episodes(feed_id):
    resp = feeds_table.get_item(
        Key={'HashID': feed_id},
        ProjectionExpression='#D',
        ExpressionAttributeNames={'#D': 'EpisodesData'}
    )

    old_episodes = []
    resp_item = resp['Item']
    raw = resp_item.get('EpisodesData')
    if not raw:
        return old_episodes

    print('Received episodes compressed data of size: {} bytes'.format(len(raw.value)))
    old_content = gzip.decompress(raw.value).decode('utf-8')  # Decompress from gzip
    old_episodes = json.loads(old_content)  # Deserialize from string to json
    return old_episodes


def _update(item):
    # Unpack fields

    feed_id = item['id']
    url = item['url']
    last_id = item['last_id']
    start = int(item['start'])
    count = int(item['count'])
    fmt = item.get('format', 'video')
    quality = item.get('quality', 'high')
    ytdl_fmt = updater.get_format(fmt, quality)

    # Playlist need special handling
    link_type = item.get('link_type')
    is_playlist = link_type == 'playlist'

    old_episodes = []

    if is_playlist:
        # Query old episodes in advance for playlist in order to compare the diff
        old_episodes = _get_episodes(feed_id)

    # Invoke youtube-dl and pull updates

    print('Updating feed {} (last id: {}, start: {}, count: {}, fmt: {}, type: {})'.format(
        feed_id, last_id, start, count, ytdl_fmt, link_type))
    new_episodes, new_last_id, dirty = updater.get_updates(start, count, url, ytdl_fmt, last_id, old_episodes)

    if new_last_id is None:
        # Sometimes youtube-dl fails to pull updates
        print('! New last id is None, retrying...')
        new_episodes, new_last_id, dirty = updater.get_updates(start, count, url, ytdl_fmt, last_id, old_episodes)

    if not dirty:
        print('No updates found for {}'.format(feed_id))
        return
    else:
        print('Found {} new episode(s) (new last id: {})'.format(len(new_episodes), new_last_id))

    # Get records from DynamoDB and decompress episodes

    if is_playlist:
        episodes = new_episodes
    else:
        old_episodes = _get_episodes(feed_id)
        episodes = new_episodes + old_episodes  # Prepand the new episodes
        if is_playlist:
            del episodes[count:]

    # Compress episodes and submit update query

    data = bytes(json.dumps(episodes), 'utf-8')
    compressed = gzip.compress(data)
    print('Sending new compressed data of size: {} bytes ({} episodes)'.format(len(compressed), len(episodes)))

    feeds_table.update_item(
        Key={
            'HashID': feed_id,
        },
        UpdateExpression='SET #episodesData = :data, #last_id = :last_id, #updated_at = :now REMOVE #episodes',
        ExpressionAttributeNames={
            '#episodesData': 'EpisodesData',
            '#episodes': 'Episodes',
            '#last_id': 'LastID',
            '#updated_at': 'UpdatedAt',
        },
        ExpressionAttributeValues={
            ':now': int(datetime.datetime.utcnow().timestamp()),
            ':last_id': new_last_id,
            ':data': compressed,
        },
    )


if __name__ == '__main__':
    print('Running updater')
    while True:
        response = sqs.receive_message(QueueUrl=sqs_url, MaxNumberOfMessages=10)
        messages = response.get('Messages')
        if not messages:
            continue

        print('=> Got {} new message(s) to process'.format(len(messages)))
        for msg in messages:
            print('-' * 64)

            body = msg.get('Body')
            receipt_handle = msg.get('ReceiptHandle')

            try:
                # Run updater
                _update(json.loads(body))
                # Delete message from SQS
                sqs.delete_message(QueueUrl=sqs_url, ReceiptHandle=receipt_handle)
                print('Done')
            except (ValueError, youtube_dl.utils.DownloadError) as e:
                print(str(e))
                # These kind of errors are not retryable, so delete message from the queue
                sqs.delete_message(QueueUrl=sqs_url, ReceiptHandle=receipt_handle)
            except Exception as e:
                print('! ERROR ({}): {}'.format(type(e), str(e)))
