package crawler

import (
	"crawler/proto"

	"net/http"
)

type Parser interface {
	Parse(resp *http.Response) *proto.ParserResult
}
