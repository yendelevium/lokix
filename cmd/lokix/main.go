package main

import (
	"log"

	"github.com/yendelevium/lokix/internal"
)

// This will act as our thread-pool
func worker(id int, jobs <-chan string, signal chan<- struct{}, queue *internal.Queue, dbClient *internal.DBClient, crawledSet *internal.CrawledSet) {
	for job := range jobs {
		// Don't rescrape things that have already been scraped
		if crawledSet.Contains(job) {
			continue
		}
		byteData := internal.FetchPage(job)
		keywords, pageHyperlinks := internal.ParseHTML(byteData, "https://en.wikipedia.org")

		log.Printf("JOB %d: URL: %s ", id, job)
		dbClient.InsertWebpage(job, keywords)
		crawledSet.Add(job)

		if queue.Empty() {
			for _, hyperlink := range pageHyperlinks {
				queue.Enqueue(hyperlink)
			}
			signal <- struct{}{}
		} else {
			for _, hyperlink := range pageHyperlinks {
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

	scheduler := internal.NewQueue()
	scheduler.Enqueue("https://en.wikipedia.org/wiki/Plant")

	crawledSet := internal.NewCrawledSet()
	scrapedCount := 0
	jobs := make(chan string, 10)
	signal := make(chan struct{}, 1)

	// Initializing the threadpool
	for i := range 10 {
		go worker(i, jobs, signal, &scheduler, &client, &crawledSet)
	}

	// Main scheduling logic
	for {
		targetURL, err := scheduler.Dequeue()
		if err != nil {
			// Wait
			<-signal
			targetURL, err = scheduler.Dequeue()

			// Termination condition (But will terminate early if the first routine to send signal doesn't have any URLs in it's HTML)
			// Even if the other workers have, I'm only checking the first signal and terminating based on that, making this flawed
			if err != nil {
				close(jobs)
				close(signal)
				break
			}
		}

		// Dispatch job to worker
		jobs <- targetURL
		scrapedCount++

		// Another ending condition -> when I readch 500 scraped URLS (imperfect again)
		// This can stop while the final (or even more) goroutine hasn't finished scraping so gotta think abt that
		if scrapedCount == 500 {
			close(jobs)
			close(signal)
			break
		}
	}

}
