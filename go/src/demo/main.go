package main

import cproto "crawler/proto"
import (
	"crawler"

	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/golang/protobuf/proto"
)

var maxBufferedFeeds = flag.Int("max_buffered_feeds", 1000000, "")
var maxOngoingCrawls = flag.Int("max_ongoing_crawls", 10, "")

func printParserResult(feed *cproto.Feed, result *cproto.ParserResult) {
	log.Println(proto.MarshalTextString(result))
}

func main() {
	flag.Parse()

	client := &http.Client{}
	scheduler := crawler.NewScheduler(*maxBufferedFeeds, *maxOngoingCrawls, printParserResult, client)
	scheduler.Run()
	for _, arg := range flag.Args() {
		file := arg
		fileContent, err := ioutil.ReadFile(file)
		if err != nil {
			log.Printf("Failed to read '%s': %s!\n", file, err)
			return
		}
		feedList := &cproto.FeedList{}
		if err := proto.UnmarshalText(string(fileContent), feedList); err != nil {
			log.Printf("Failed to parse '%s': %s!\n", file, err)
			return
		}
		for _, feed := range feedList.Feed {
			log.Println(feed)
			scheduler.AddFeed(feed)
		}
	}
	scheduler.CloseAndWait()
}
