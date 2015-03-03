package crawler

import (
	"crawler/proto"
	"log"
	"net/http"
)

type ParserManager struct {
	parserMap map[string]Parser
}

func NewParserManager() *ParserManager {
	parerMap := map[string]Parser{
		"amazon": &GoQueryProductParser{
			// nameElQuery
			"#title",
			// imgElQuery
			"#imgTagWrapperId > img",
		},
		"amazon_list": &GoQueryProductListParser{
			// itemQuery
			".s-item-container",
			// linkElQuery
			".a-row:nth-child(2) > .a-row:first-child > .a-link-normal",
			// nameElQuery
			".a-row:nth-child(2) > .a-row:first-child > .a-link-normal",
			// imgElQuery
			".a-row:first-child > .a-column > .a-section > .a-link-normal > img",
		},
	}
	return &ParserManager{parerMap}
}

func (pm *ParserManager) Parse(resp *http.Response, parserName string) *proto.ParserResult {
	parser, found := pm.parserMap[parserName]
	if !found {
		log.Panicf("Unkown parser '%s'!\n", parserName)
	}
	return parser.Parse(resp)
}
