import function
import unittest

TEST_URL = 'https://www.youtube.com/user/CNN/videos'


class TestUpdater(unittest.TestCase):
    #@unittest.skip('heavy test, run manually')
    def test_get_50(self):
        resp = function.handler({
            'url': 'https://www.youtube.com/channel/UCd6MoB9NC6uYN2grvUNT-Zg',
            'start': 1,
            'count': 50,
            'format': 'audio',
            'quality': 'low',
        }, None)
        self.assertEqual(len(resp['Episodes']), 50)
        self.assertEqual(resp['Episodes'][0]['ID'], resp['LastID'])
