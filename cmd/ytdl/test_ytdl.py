import ytdl
import unittest


class TestYtdl(unittest.TestCase):
    def test_resolve(self):
        self.assertIsNotNone(
            ytdl._resolve('https://youtube.com/watch?v=ygIUF678y40', {'format': 'video', 'quality': 'low', 'provider': 'youtube'}))
        self.assertIsNotNone(
            ytdl._resolve('https://youtube.com/watch?v=WyaEiO4hyik', {'format': 'audio', 'quality': 'high', 'provider': 'youtube'}))

    def test_vimeo(self):
        self.assertIsNotNone(
            ytdl._resolve('https://vimeo.com/237715420', {'format': 'video', 'quality': 'low', 'provider': 'vimeo'}))
        self.assertIsNotNone(
            ytdl._resolve('https://vimeo.com/275211960', {'format': 'video', 'quality': 'high', 'provider': 'vimeo' })
        )
