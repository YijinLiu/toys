#!/usr/bin/python

from Queue import Empty
from Queue import Queue
from abc import ABCMeta
from abc import abstractmethod
import logging
import re
import threading
import urllib2

from pyquery import PyQuery

from parser_pb2 import ParserResult

class Parser:
	__metaclass__ = ABCMeta
	
	@abstractmethod
	def Parse(self, html_data):
		pass

_space_re = re.compile('\\s+')

def _SanitizeTextContent(text_content):
	return re.sub(_space_re, ' ', text_content.strip())

def _GetText(el):
	return _SanitizeTextContent(PyQuery(el).text())

def _GetImageSrc(image_el):
	# Handle Amazon specially.
	if image_el.attrib.has_key('data-old-hires'):
		return image_el.attrib['data-old-hires']
	if image_el.attrib.has_key('src'):
		return image_el.attrib['src']
	return None

def _GetLinkHref(link_el):
	# Handle Amazon specially.
	if link_el.attrib.has_key('href'):
		return link_el.attrib['href']
	return None

class SoupBasedProductParser:
	def __init__(self, name_el_query, image_el_query):
		self._name_el_query = name_el_query
		self._image_el_query = image_el_query

	def Parse(self, html_data):
		parser_result = ParserResult()
		product = parser_result.product
		doc = PyQuery(html_data)

		# Get name.
		name_el = doc(self._name_el_query)
		if not name_el:
			logging.info('Failed to find name element!')
			return None
		product.name = _GetText(name_el[0])

		# Get image.
		image_el = doc(self._image_el_query)
		if not image_el:
			logging.info('Failed to find image element!')
			return None
		image_src = _GetImageSrc(image_el[0])
		if not image_src:
			logging.info('Failed to find image src!')
			return None
		product.image_url = image_src.strip()
		return parser_result

Parser.register(SoupBasedProductParser)

class SoupBasedProductListParser:
	def __init__(self, item_query, link_el_query, name_el_query, image_el_query):
		self._item_query = item_query
		self._link_el_query = link_el_query
		self._name_el_query = name_el_query
		self._image_el_query = image_el_query

	def Parse(self, html_data):
		parser_result = ParserResult()
		product_list = parser_result.product_list
		doc = PyQuery(html_data)
		for productEl in doc.items(self._item_query):
			link_el = productEl.find(self._link_el_query)
			if not link_el:
				logging.info('Failed to find link element!')
				continue
			url = _GetLinkHref(link_el[0])
			if not url:
				logging.info('Failed to find href!')
				continue

			name_el = productEl.find(self._name_el_query)
			if not name_el:
				logging.info('Failed to find name element!')
				continue
			name = _GetText(name_el[0])

			image_el = productEl.find(self._image_el_query)
			if not image_el:
				logging.info('Failed to find image element!')
				continue
			image_url = _GetImageSrc(image_el[0])
			if not image_url:
				logging.info('Failed to find image src!')
				continue

			product = product_list.product.add()
			product.url = url.strip()
			product.name = name
			product.image_url = image_url.strip()
		return parser_result

Parser.register(SoupBasedProductListParser)

class ParserManager:
	def __init__(self):
		self._parser_dict = {
			'amazon': SoupBasedProductParser('#title', '#imgTagWrapperId > img'),
			'amazon_list': SoupBasedProductListParser('.s-item-container',
								  '.a-row > .a-row > .a-link-normal',
								  '.a-row > .a-row > .a-link-normal',
								  '.a-row > .a-column > .a-section > .a-link-normal > img')
		}

	def Parse(self, html_data, parser_name):
		if not self._parser_dict.has_key(parser_name):
			logging.info('Unknown parser: %s!' % parser_name)
			return None
		return self._parser_dict[parser_name].Parse(html_data)

class CrawlThread(threading.Thread):
	def __init__ (self, order, scheduler):
		threading.Thread.__init__(self)
		self.daemon = True
		self._order = order
		self._scheduler = scheduler

	def run(self):
		self._scheduler.Crawl(self._order)

class Scheduler(object):
	def __init__(self, max_ongoing_crawls, buffer_size, output_cb):
		self._max_ongoing_crawls = max_ongoing_crawls
		self._input_queue = Queue(buffer_size)
		self._output_cb = output_cb
		self._closing = False
		self._parser_manager = ParserManager()

	def AddFeed(self, feed):
		assert not self._closing
	 	self._input_queue.put(feed, True)

	def Crawl(self, i):
		# Get feed.
		while True:
			feed = None
			if self._closing:
				if self._input_queue.empty():
					return
				try:
					feed = self._input_queue.get(False)
				except Empty:
					continue
			else:
				# We need timeout because the main thread might call Wait after this thread call get.
				try:
					feed = self._input_queue.get(True, 1)
				except Empty:
					continue

   			# Crawl and parse
			html_data = None
			try:
				headers = {"User-Agent":"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/39.0.2171.65 Chrome/39.0.2171.65 Safari/537.36"}
				request = urllib2.Request(feed.url, None, headers)
				html_data = urllib2.urlopen(request).read()
			except Exception as e:
				logging.info("Error in crawling %s: %s." % (feed.url, e))
			result = None
			if html_data:
				result = self._parser_manager.Parse(html_data, feed.parser)
			self._output_cb(feed, result)
	
			self._input_queue.task_done()
	
	def Run(self):
		# Start max_ongoing_crawls threads.
		for i in range(self._max_ongoing_crawls):
			crawl_thread = CrawlThread(i, self)
			crawl_thread.start()

	def Wait(self):
		self._closing = True
		self._input_queue.join()
