import updater
import unittest

TEST_URL = 'https://www.youtube.com/user/CNN/videos'


class TestUpdater(unittest.TestCase):
    def test_get_updates(self):
        kinds = [
            updater._get_format({'Format': 'video', 'Quality': 'high'}),
            updater._get_format({'Format': 'video', 'Quality': 'low'}),
            updater._get_format({'Format': 'audio', 'Quality': 'high'}),
            updater._get_format({'Format': 'audio', 'Quality': 'low'}),
        ]
        for kind in kinds:
            with self.subTest(kind):
                feed, items, _ = updater._get_updates(1, 1, TEST_URL, kind)
                self.assertIsNotNone(feed)
                self.assertIsNotNone(items)

    def test_get_change_list(self):
        feed, items, last_id = updater._get_updates(1, 5, TEST_URL, 'worst[ext=mp4]')

        self.assertEqual(len(items), 5)
        self.assertEqual(items[0]['ID'], last_id)
        test_last_id = items[2]['ID']
        self.assertIsNotNone(test_last_id)

        feed, items, last_id = updater._get_updates(1, 5, TEST_URL, 'worst[ext=mp4]', test_last_id)
        self.assertEqual(len(items), 2)
        self.assertEqual(items[0]['ID'], last_id)

    def test_last_id(self):
        feed, items, last_id = updater._get_updates(1, 1, TEST_URL, 'worstaudio')
        self.assertEqual(len(items), 1)
        self.assertEqual(items[0]['ID'], last_id)

    @unittest.skip('heavy test, run manually')
    def test_get_50(self):
        _, items, last_id = updater.handler({
            'url': 'https://www.youtube.com/channel/UCd6MoB9NC6uYN2grvUNT-Zg',
            'start': 1,
            'count': 50,
            'kind': 'best[ext=mp4]',
        }, None)
        self.assertEqual(len(items), 50)
        self.assertEqual(items[0]['ID'], last_id)

    @unittest.skip
    def test_update_feed(self):
        updater._update_feed('86qZ')
