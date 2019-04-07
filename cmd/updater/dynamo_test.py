import dynamo
import unittest

TEST_FEED = '86qZ'


class TestUpdater(unittest.TestCase):
    @unittest.skip
    def test_update_feed(self):
        dynamo._update_feed(TEST_FEED)
