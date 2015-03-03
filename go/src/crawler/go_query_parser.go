package crawler

import cproto "crawler/proto"

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/protobuf/proto"
)

type GoQueryProductParser struct {
	nameElQuery string
	imgElQuery  string
}

var spaceRe = regexp.MustCompile("\\s+")

func sanitizeTextContent(textContent string) string {
	return spaceRe.ReplaceAllLiteralString(strings.TrimSpace(textContent), " ")
}

func getImageUrl(imgEl *goquery.Selection) string {
	amazonImg, exists := imgEl.Attr("data-old-hires")
	if exists {
		return amazonImg
	}
	imageSrc, exists := imgEl.Attr("src")
	if exists {
		return imageSrc
	}
	return ""
}

func (p *GoQueryProductParser) Parse(resp *http.Response) *cproto.ParserResult {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Printf("Failed to parse: %s!\n", err)
		return nil
	}
	product := &cproto.Product{}

	// Get name.
	nameEl := doc.Find(p.nameElQuery)
	if nameEl.Size() == 0 {
		log.Println("Failed to find name element!")
		return nil
	}
	product.Name = proto.String(sanitizeTextContent(nameEl.Text()))

	// Get image url.
	imgEl := doc.Find(p.imgElQuery)
	if imgEl.Size() == 0 {
		log.Println("Failed to find img element!")
		return nil
	}
	imageUrl := getImageUrl(imgEl)
	if len(imageUrl) == 0 {
		log.Println("Failed to find image src!")
		return nil
	}
	product.ImageUrl = proto.String(sanitizeTextContent(imageUrl))
	result := &cproto.ParserResult{}
	result.Product = product
	return result
}

type GoQueryProductListParser struct {
	itemQuery   string
	linkElQuery string
	nameElQuery string
	imgElQuery  string
}

func (p *GoQueryProductListParser) Parse(resp *http.Response) *cproto.ParserResult {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Printf("Failed to parse: %s!\n", err)
		return nil
	}
	productList := &cproto.ProductList{}
	doc.Find(p.itemQuery).Each(func(i int, s *goquery.Selection) {
		product := &cproto.Product{}

		// Get url.
		linkEl := s.Find(p.linkElQuery)
		if linkEl.Size() == 0 {
			log.Println("Failed to find link element!")
			return
		}
		url, exists := linkEl.Attr("href")
		if !exists {
			log.Println("Failed to find href!")
			return
		}
		product.Url = proto.String(url)

		// Get name.
		nameEl := s.Find(p.nameElQuery)
		if nameEl.Size() == 0 {
			log.Println("Failed to find name element!")
			return
		}
		product.Name = proto.String(sanitizeTextContent(nameEl.Text()))

		// Get image url.
		imgEl := s.Find(p.imgElQuery)
		if imgEl.Size() == 0 {
			log.Println("Failed to find img element!")
			return
		}
		imageUrl, exists := imgEl.Attr("src")
		if !exists {
			log.Println("Failed to find src!")
			return
		}
		product.ImageUrl = proto.String(imageUrl)

		productList.Product = append(productList.Product, product)
	})
	return &cproto.ParserResult{ProductList: productList}
}
