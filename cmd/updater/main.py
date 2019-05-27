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

    # Invoke youtube-dl and pull updates

    print('Updating feed {} (last id: {}, start: {}, count: {}, fmt: {})'.format(
        feed_id, last_id, start, count, ytdl_fmt))
    _, new_episodes, new_last_id = updater.get_updates(start, count, url, ytdl_fmt, last_id)

    if new_last_id is None:
        # Sometimes youtube-dl fails to pull updates
        print('! New last id is None, retrying...')
        _, new_episodes, new_last_id = updater.get_updates(start, count, url, ytdl_fmt, last_id)

    if new_last_id == last_id:
        print('No updates found for {}'.format(feed_id))
        return
    else:
        print('Found {} new episode(s) (new last id: {})'.format(len(new_episodes), new_last_id))

    # Get record and DynamoDB and decompress episodes

    resp = feeds_table.get_item(
        Key={'HashID': feed_id},
        ProjectionExpression='#D',
        ExpressionAttributeNames={'#D': 'EpisodesData'}
    )

    old_episodes = []
    resp_item = resp['Item']
    raw = resp_item.get('EpisodesData')
    if raw:
        print('Received episodes compressed data of size: {} bytes'.format(len(raw.value)))
        old_content = gzip.decompress(raw.value).decode('utf-8')  # Decompress from gzip
        old_episodes = json.loads(old_content)  # Deserialize from string to json

    episodes = new_episodes + old_episodes  # Prepand new episodes to the list
    if len(episodes) > count:
        del episodes[count:]  # Truncate list

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
