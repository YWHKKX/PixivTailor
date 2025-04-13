package crawler

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"

	"github.com/andybalholm/brotli"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

func decodeToUTF8(res *http.Response, body []byte) []byte {
	encoding, _, _ := charset.DetermineEncoding(body, res.Header.Get("Content-Type"))
	decoder := encoding.NewDecoder()

	utf8Body, _, _ := transform.Bytes(decoder, body)
	return utf8Body
}

func decodeZip(res *http.Response) []byte {
	var body []byte
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ := gzip.NewReader(res.Body)
		body, _ = ioutil.ReadAll(reader)
	case "br":
		reader := brotli.NewReader(res.Body)
		body, _ = ioutil.ReadAll(reader)
	default:
		body, _ = ioutil.ReadAll(res.Body)
	}
	return body
}
