package main

import (
	"log"
	"time"

	"github.com/yendelevium/lokix/internal"
	"github.com/yendelevium/lokix/internal/collections"
)

// This will act as our thread-pool
func worker(id int, jobs <-chan string, signal chan<- struct{}, queue *collections.Queue, dbClient *internal.DBClient, crawledSet *collections.CrawledSet) {
	for job := range jobs {
		// Don't rescrape things that have already been scraped
		if crawledSet.Contains(job) {
			continue
		}
		byteData := internal.FetchPage(job)
		keywords, pageHyperlinks := internal.ParseHTML(byteData, "https://en.wikipedia.org")

		// log.Printf("JOB %d: URL: %s ", id, job)
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

	scheduler := collections.NewQueue()
	scheduler.Enqueue("https://en.wikipedia.org/wiki/Plant")

	crawledSet := collections.NewCrawledSet()
	jobs := make(chan string, 10)
	signal := make(chan struct{}, 1)

	// Initializing the threadpool
	for i := range 10 {
		go worker(i, jobs, signal, &scheduler, &client, &crawledSet)
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

		// Another ending condition -> when I reach 2000 scraped URLS (imperfect again)
		// This can stop while the final (or even more) goroutine hasn't finished scraping so gotta think abt that
		if crawledSet.Total() == 2000 {
			close(jobs)
			close(signal)
			break
		}
	}

	// Stop the stats
	ticker.Stop()
	done <- true
	log.Printf("Total Pages Crawled: %d, Total Queued: %d", crawledSet.Total(), scheduler.TotalQueued())
}
