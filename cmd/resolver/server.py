import resolver

from sanic import Sanic, response
from sanic.exceptions import ServerError, InvalidUsage

app = Sanic()


@app.get('/download/<feed_id>/<video_id>')
async def download(req, feed_id, video_id):
    try:
        redirect_url = resolver.download(feed_id, video_id)
        return response.redirect(redirect_url)
    except resolver.InvalidUsage:
        raise InvalidUsage()
    except resolver.QuotaExceeded:
        raise ServerError('Too many requests. Daily limit is 1000. Consider upgrading account to get unlimited access.',
                          status_code=429)


@app.get('/ping')
async def ping(req):
    return response.text('pong')


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)
