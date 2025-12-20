package main

import (
	"log"

	"github.com/yendelevium/lokix/internal"
)

func main() {
	log.Println("BYE, lokix")
	internal.FetchPage("https://google.com")
}
