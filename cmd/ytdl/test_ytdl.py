import ytdl
import unittest


class TestYtdl(unittest.TestCase):
    def test_resolve(self):
        self.assertIsNotNone(ytdl._resolve('https://youtube.com/watch?v=ygIUF678y40', {'format': 'video', 'quality': 'low'}))
        self.assertIsNotNone(ytdl._resolve('https://youtube.com/watch?v=WyaEiO4hyik', {'format': 'audio', 'quality': 'high'}))
