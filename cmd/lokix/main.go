package main

import (
	"log"

	"github.com/yendelevium/lokix/internal"
)

func main() {
	log.Println("BYE, lokix")
	byteData := internal.FetchPage("https://en.wikipedia.org/wiki/Plant")
	internal.ParseHTML(byteData, "https://en.wikipedia.org")
}
