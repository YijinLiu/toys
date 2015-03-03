package crawler

import cproto "crawler/proto"

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

type Scheduler struct {
	maxOngoingCrawls int
	outputCB         func(feed *cproto.Feed, result *cproto.ParserResult)
	inputQueue       chan *cproto.Feed
	crawlQueue       chan *cproto.Feed
	client           *http.Client
	parserManager    *ParserManager
	waitGroup        *sync.WaitGroup
}

// Set maxBufferedFeeds, and maxOngoingCrawls to modest small number, to avoid OOM.
func NewScheduler(maxBufferedFeeds, maxOngoingCrawls int,
	outputCB func(feed *cproto.Feed, result *cproto.ParserResult),
	client *http.Client) *Scheduler {
	scheduler := &Scheduler{
		maxOngoingCrawls, outputCB,
		make(chan *cproto.Feed, maxBufferedFeeds),
		make(chan *cproto.Feed, maxOngoingCrawls),
		client, NewParserManager(), &sync.WaitGroup{},
	}
	return scheduler
}

func (s *Scheduler) crawl(workerId int) {
	for {
		feed, more := <-s.crawlQueue
		if !more {
			break
		}
		log.Printf("Crawling %s ...\n", feed.GetUrl())
		resp, err := s.client.Get(feed.GetUrl())
		if err != nil {
			log.Printf("Failed to crawl '%s': %s.\n", feed.Url, err)
			s.outputCB(feed, nil)
			continue
		}
		crawlActivity := &cproto.CrawlActivity{
			Timestamp:  proto.Int64(time.Now().Unix()),
			StatusCode: proto.Int32(int32(resp.StatusCode)),
		}
		feed.CrawlActivity = append(feed.CrawlActivity, crawlActivity)
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("Non 200 status code for '%s': %s.\n", feed.Url, resp.Status)
			s.outputCB(feed, nil)
			continue
		}
		s.outputCB(feed, s.parserManager.Parse(resp, feed.GetParser()))

	}
	log.Printf("Worker %d exiting ...", workerId)
	s.waitGroup.Done()

}

// May block if there are too many buffered feeds.
func (s *Scheduler) AddFeed(feed *cproto.Feed) {
	s.inputQueue <- feed
}

// Crawls all feeds one time.
func (s *Scheduler) Run() {
	s.waitGroup.Add(s.maxOngoingCrawls + 1)

	// Start maxOngoingCrawls go routings.
	for i := 0; i < s.maxOngoingCrawls; i++ {
		go s.crawl(i)
	}

	go func() {
		for {
			feed, more := <-s.inputQueue
			if !more {
				break
			}
			// TODO: Add per host limit.
			s.crawlQueue <- feed
		}
		close(s.crawlQueue)
		s.waitGroup.Done()
	}()
}

// Closes all the worker routines.
func (s *Scheduler) CloseAndWait() {
	close(s.inputQueue)
	s.waitGroup.Wait()
}
