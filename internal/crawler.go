package internal

import (
	"io"
	"log"
	"net/http"
)

func FetchPage(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch URL: %s", err.Error())
		return []byte{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %s", err.Error())
		return []byte{}
	}

	log.Println(string(body))
	return body
}
