#!/usr/bin/python

import logging
import sys

from google.protobuf import text_format

from crawler import Scheduler
from feed_pb2 import FeedList

def PrintResult(feed, result):
	print feed
	print result

logger = logging.getLogger()
logger.setLevel(logging.DEBUG)
handler = logging.StreamHandler()
logger.addHandler(handler)
scheduler = Scheduler(10, 1000000, PrintResult)
scheduler.Run()

for i in range(1, len(sys.argv)):
	feed_list = FeedList()
	text_format.Merge(open(sys.argv[i]).read(), feed_list)
	for feed in feed_list.feed:
		scheduler.AddFeed(feed)

scheduler.Wait()
