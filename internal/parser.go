package internal

import (
	"bytes"
	"io"
	"log"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

// BLogs -> https://www.zenrows.com/blog/golang-html-parser#prerequisites
// https://medium.com/@datajournal/parse-html-in-golang-83c882576a0a
// https://zetcode.com/golang/net-html/

// Can add a lot more words to ignore but for now this is enough
var ignoreWords []string = []string{"a", "an", "the", ",", ";", ":", "-", "_", "!", "[", "]", "{", "}", "(", ")", ".", "\n"}
var ignoreTags []string = []string{"script", "style"}

// This ParseHTML function is tailored to parse Wikipedia Pages (removing all unnecessary headers and stuff, with page of contents)
// To make it more general purpose remove the startParsing logic
func ParseHTML(htmlData []byte, sourceURL string) ([]string, []string) {
	if len(htmlData) == 0 {
		return []string{}, []string{}
	}
	byteReader := bytes.NewReader(htmlData)
	tokenizer := html.NewTokenizer(byteReader)

	// This is important to filter out unimportant script and style tags
	// But now with the startParsing logic we probably don't need it as most of it will be ignored anyways
	previousStartTag := "html"
	keywords := make([]string, 1000)
	keywordIdx := 0

	pageHyperlinks := make([]string, 50)
	pageHyperlinkIdx := 0

	startParsing := false

	// TODO: Early exit if both URL limit and keyword Limit have hit max
	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()

		switch tokenType {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return keywords, pageHyperlinks
			}
			log.Printf("Error: %v", tokenizer.Err())
			return keywords, pageHyperlinks

		// Wrapping the 2nd switch_case inside html.StartTagToken as otherwise
		// token.Data activates twice - once for opening and one for closing
		case html.StartTagToken:
			// Storing the token.Data to later skip any script or style tags
			previousStartTag = token.Data
			switch token.Data {
			case "div":
				for _, attr := range token.Attr {
					if attr.Key == "id" && attr.Val == "bodyContent" {
						startParsing = true
					}
				}

			case "a":
				if !startParsing {
					continue
				}
				for _, attr := range token.Attr {
					// Need to sanitize the hyperlinks -> no query params in the link + add protocol + domainname (if missing)
					// This implementation is tailored to wikipedia
					// Use the net/url pkg for this
					if attr.Key == "href" {
						if pageHyperlinkIdx >= len(pageHyperlinks) {
							continue
						}

						parsedURL, err := url.Parse(attr.Val)
						if err != nil {
							log.Printf("Malformed URL: %v", err)
						}
						pageHyperlinks[pageHyperlinkIdx] = sourceURL + parsedURL.Path
						pageHyperlinkIdx++
					}
				}
			}

		case html.TextToken:
			if !startParsing {
				continue
			}
			// To skip script and style tags that take up the word count
			if slices.Contains(ignoreTags, previousStartTag) {
				// log.Println("Skip", previousStartTag)
				continue
			}

			words := strings.Split(token.Data, " ")
			for _, word := range words {
				// Removing filler words
				if slices.Contains(ignoreWords, word) {
					continue
				}
				// Max count hit; Can't break as we might find URLs that come later on in the HTML block
				if keywordIdx >= len(keywords) {
					continue
				}

				// Make everything lowercase for uniformity
				sanitizedWord := strings.ToLower(strings.TrimSpace(word))
				if sanitizedWord != "" {
					keywords[keywordIdx] = word
					keywordIdx++
				}
			}
		}
	}
}
