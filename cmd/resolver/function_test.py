import function as ytdl
import unittest


class TestYtdl(unittest.TestCase):
    def test_resolve(self):
        self.assertIsNotNone(
            ytdl._resolve('https://youtube.com/watch?v=ygIUF678y40',
                          {'format': 'video', 'quality': 'low', 'provider': 'youtube'}))
        self.assertIsNotNone(
            ytdl._resolve('https://youtube.com/watch?v=WyaEiO4hyik',
                          {'format': 'audio', 'quality': 'high', 'provider': 'youtube'}))
        self.assertIsNotNone(
            ytdl._resolve('https://youtube.com/watch?v=5mjUF2j9dgA',
                          {'format': 'video', 'quality': 'low', 'provider': 'youtube'})
        )
        self.assertIsNotNone(
            ytdl._resolve('https://www.youtube.com/watch?v=2nH7xAMqD2g',
                          {'format': 'video', 'quality': 'high', 'provider': 'youtube'})
        )

    def test_vimeo(self):
        self.assertIsNotNone(
            ytdl._resolve('https://vimeo.com/237715420', {'format': 'video', 'quality': 'low', 'provider': 'vimeo'}))
        self.assertIsNotNone(
            ytdl._resolve('https://vimeo.com/275211960', {'format': 'video', 'quality': 'high', 'provider': 'vimeo'})
        )

    def test_youtube_resolve_audio(self):
        self.assertIsNotNone(
            ytdl._resolve('https://youtube.com/watch?v=UMrb1tG38w8',
                          {'format': 'audio', 'quality': 'high', 'provider': 'youtube'})
        )

    def test_get_ids(self):
        feed_id, video_id = ytdl._get_ids('/download/1/2.xml')
        self.assertEqual('1', feed_id)
        self.assertEqual('2', video_id)

        feed_id, video_id = ytdl._get_ids('/download/1/2')
        self.assertEqual('1', feed_id)
        self.assertEqual('2', video_id)

    def test_get_invalid_ids(self):
        with self.assertRaises(ytdl.InvalidUsage):
            ytdl._get_ids(None)

        with self.assertRaises(ytdl.InvalidUsage):
            ytdl._get_ids('')

        with self.assertRaises(ytdl.InvalidUsage):
            ytdl._get_ids('/download/')

        with self.assertRaises(ytdl.InvalidUsage):
            ytdl._get_ids('/download/1')

        with self.assertRaises(ytdl.InvalidUsage):
            ytdl._get_ids('/download/1/')

        with self.assertRaises(ytdl.InvalidUsage):
            ytdl._get_ids('/download/1/2/3')


class TestDynamo(unittest.TestCase):
    def test_metadata(self):
        item = ytdl._get_metadata('86qZ')
        self.assertIsNotNone(item)
        self.assertIsNotNone(item['format'])
        self.assertIsNotNone(item['quality'])
        self.assertIsNotNone(item['provider'])

    def test_counter(self):
        counter = ytdl._update_resolve_counter('86qZ')
        self.assertEqual(counter, 1)
        counter = ytdl._update_resolve_counter('86qZ')
        self.assertEqual(counter, 2)

    def test_download(self):
        url = ytdl.download('86qZ', '7XJYLq3gviY')
        self.assertIsNotNone(url)

    def test_quota_check(self):
        with self.assertRaises(ytdl.QuotaExceeded):
            ytdl.download('xro548QlJ', 'j51NFs0bZ9c')
