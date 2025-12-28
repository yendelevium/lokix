package main

import (
	"log"
	"time"

	"github.com/yendelevium/lokix/internal"
	"github.com/yendelevium/lokix/internal/collections"
)

type ParseJob struct {
	byteData []byte
	url      string
}

const WORKERS = 20

// These workers will act as our threadpool(s)
func fetchWorker(id int, fetchURLs <-chan string, parseData chan<- ParseJob) {
	for url := range fetchURLs {
		byteData := internal.FetchPage(url)
		// log.Printf("FETCH JOB %d: URL: %s ", id, url)
		if len(byteData) == 0 {
			// If no data recieved from the URL (maybe a 404), wait for the next job
			continue
		}

		parseData <- ParseJob{
			byteData,
			url,
		}
	}
}

func parseWorker(id int, parseData <-chan ParseJob, queue *collections.Queue, dbClient *internal.DBClient, crawledSet *collections.CrawledSet) {
	for data := range parseData {
		keywords, pageHyperlinks := internal.ParseHTML(data.byteData, "https://en.wikipedia.org")

		// log.Printf("PARSE JOB %d: URL: %s ", id, data.url)
		dbClient.InsertWebpage(data.url, keywords)
		crawledSet.Add(data.url)

		// Add the page-links to the scheduler
		for _, hyperlink := range pageHyperlinks {
			if hyperlink == "" {
				// The page didn't have min 50 URLS, so the remaining are empty
				continue
			}

			// Add only if not already crawled
			if !crawledSet.Contains(hyperlink) {
				queue.Enqueue(hyperlink)
			}
		}

	}
}

func main() {
	log.Println("BYE, lokix")

	// Connect to DB
	client := internal.ConnectMongo()
	defer client.Disconnect()

	scheduler := collections.NewQueue()
	scheduler.Enqueue("https://en.wikipedia.org/wiki/Plant")

	crawledSet := collections.NewCrawledSet()
	fetchURLs := make(chan string, WORKERS)
	parseData := make(chan ParseJob, WORKERS)

	// Initializing the threadpool(s)
	for i := range WORKERS {
		go fetchWorker(i, fetchURLs, parseData)
	}
	for i := range WORKERS {
		go parseWorker(i, parseData, &scheduler, &client, &crawledSet)
	}

	// Crawler Stats -> Every 10 seconds
	log.Printf("Starting Crawling")
	ticker := time.NewTicker(10 * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				log.Printf("Total Pages Crawled: %d, Total Queued: %d", crawledSet.Total(), scheduler.TotalQueued())
			}
		}
	}()

	// Dispatch first job to enqueue additional URLS
	targetURL, _ := scheduler.Dequeue()
	fetchURLs <- targetURL

	for scheduler.Empty() {
		// Wait till queue is filled by the first seed
		// TODO: This is busy waiting :( would like to avoid this if I can
	}

	// Main scheduling logic
	for crawledSet.Total() <= 2000 && !scheduler.Empty() {
		targetURL, _ = scheduler.Dequeue()
		if crawledSet.Contains(targetURL) {
			// Don't rescrape existing URLs
			continue
		}

		// Dispatch jobs
		fetchURLs <- targetURL
	}

	// Stop the stats
	ticker.Stop()
	done <- true
	log.Printf("Total Pages Crawled: %d, Total Queued: %d", crawledSet.Total(), scheduler.TotalQueued())

	// Creating the inverted index
	err := client.CreateInvertedIndex()
	if err != nil {
		log.Fatalf("Failed to create inverted index, %v", err)
	}

	log.Println("Showing results when searching for the word 'eukaryote'")
	err = client.Search("eukaryote")
	if err != nil {
		log.Printf("Failed to search index, %v", err)
	}
}
