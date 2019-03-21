import function
import unittest


class TestUpdater(unittest.TestCase):
    def test_get_updates(self):
        kinds = ['video_high', 'video_low', 'audio_high', 'audio_low']
        for kind in kinds:
            with self.subTest(kind):
                result = function._get_updates(1, 2, 'https://www.youtube.com/user/CNN/videos', kind)
                self.assertIsNotNone(result['feed'])
                self.assertIsNotNone(result['items'])

    def test_get_change_list(self):
        result = function._get_updates(1, 5, 'https://www.youtube.com/user/CNN/videos', 'video_low')
        self.assertEqual(len(result['items']), 5)
        self.assertEqual(result['items'][0]['id'], result['last_id'])
        last_id = result['items'][2]['id']
        self.assertIsNotNone(last_id)
        result = function._get_updates(1, 5, 'https://www.youtube.com/user/CNN/videos', 'video_low', last_id)
        self.assertEqual(len(result['items']), 2)
        self.assertEqual(result['items'][0]['id'], result['last_id'])
