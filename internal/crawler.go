package internal

import (
	"io"
	"log"
	"net/http"
)

func FetchPage(url string) []byte {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	// Some sites need a user-agent to allow crawling
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to fetch: %s, URL: %s", err.Error(), url)
		return []byte{}
	}

	if resp.StatusCode != http.StatusOK {
		// Couldn't fetch URL, so return empty body
		return []byte{}
	}

	defer resp.Body.Close()
	// log.Println("Recvd Data", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %s", err.Error())
		return []byte{}
	}

	return body
}
