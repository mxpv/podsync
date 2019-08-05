import updater
import unittest

TEST_URL = 'https://www.youtube.com/user/CNN/videos'


class TestUpdater(unittest.TestCase):
    def test_get_updates(self):
        kinds = [
            updater.get_format('video', 'high'),
            updater.get_format('video', 'low'),
            updater.get_format('audio', 'high'),
            updater.get_format('audio', 'low'),
        ]
        for kind in kinds:
            with self.subTest(kind):
                items, last_id, _ = updater.get_updates(1, 1, TEST_URL, kind)
                self.assertIsNotNone(items)
                self.assertIsNotNone(last_id)

    def test_get_change_list(self):
        items, last_id, _ = updater.get_updates(1, 5, TEST_URL, 'worst[ext=mp4]')

        self.assertEqual(len(items), 5)
        self.assertEqual(items[0]['ID'], last_id)
        test_last_id = items[2]['ID']
        self.assertIsNotNone(test_last_id)

        items, last_id, _ = updater.get_updates(1, 5, TEST_URL, 'worst[ext=mp4]', test_last_id)
        self.assertEqual(len(items), 2)
        self.assertEqual(items[0]['ID'], last_id)

    def test_last_id(self):
        items, last_id, _ = updater.get_updates(1, 1, TEST_URL, 'worstaudio')
        self.assertEqual(len(items), 1)
        self.assertEqual(items[0]['ID'], last_id)

    def test_get_title_issue33(self):
        url = 'https://youtube.com/channel/UC9-y-6csu5WGm29I7JiwpnA'
        items, _, _ = updater.get_updates(1, 1, url, 'best[ext=mp4]')
        for item in items:
            self.assertNotEqual('_', item.get('Title'))
