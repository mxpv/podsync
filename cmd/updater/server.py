import updater
from sanic import Sanic, response
from sanic.exceptions import InvalidUsage

app = Sanic()


@app.get('/update')
async def update(req):
    url = req.args.get('url', None)
    start = req.args.get('start', 1)
    count = req.args.get('count', updater.DEFAULT_PAGE_SIZE)

    # Last seen video ID
    last_id = req.args.get('last_id', None)

    # Detect item format
    fmt = req.args.get('format', 'video')
    quality = req.args.get('quality', 'high')
    ytdl_fmt = updater.get_format(fmt, quality)

    try:
        _, episodes, new_last_id = updater.get_updates(start, count, url, ytdl_fmt, last_id)
        return response.json({
            'last_id': new_last_id,
            'episodes': episodes,
        })
    except ValueError:
        raise InvalidUsage()


@app.get('/ping')
async def ping(req):
    return response.text('pong')


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)
