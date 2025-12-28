package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/yendelevium/lokix/internal"
)

type node struct {
	val  string
	next *node
}

type Queue struct {
	head *node
	tail *node
	mu   *sync.Mutex
}

// Implement the Queue
func (q *Queue) Enqueue(item string) {
	newNode := node{
		val:  item,
		next: nil,
	}

	if q.head == nil {
		q.head = &newNode
		q.tail = &newNode
	} else {
		q.tail.next = &newNode
		q.tail = q.tail.next
	}
}

func (q *Queue) Dequeue() (string, error) {
	if q.head == nil {
		return "", fmt.Errorf("Queue is EMPTY!")
	}
	top := q.head.val
	q.head = q.head.next
	if q.head == nil {
		q.tail = nil
	}
	return top, nil

}

// This will act as our thread-pool
func worker(id int, jobs <-chan string, signal chan<- struct{}, queue *Queue, dbClient *internal.DBClient) {
	for job := range jobs {
		byteData := internal.FetchPage(job)
		keywords, pageHyperlinks := internal.ParseHTML(byteData, "https://en.wikipedia.org")

		log.Printf("JOB %d: URL: %s DATA: %v", id, job, keywords)
		dbClient.Mu.Lock()
		// DB Stuff
		dbClient.Mu.Unlock()
		queue.mu.Lock()
		if queue.head == nil {
			for _, hyperlink := range pageHyperlinks {
				queue.Enqueue(hyperlink)
			}
			signal <- struct{}{}
		} else {
			for _, hyperlink := range pageHyperlinks {
				queue.Enqueue(hyperlink)
			}
		}
		queue.mu.Unlock()
	}
}

func main() {
	log.Println("BYE, lokix")

	// Connect to DB
	client := internal.ConnectMongo()

	seed := node{
		val:  "https://en.wikipedia.org/wiki/Plant",
		next: nil,
	}

	scheduler := Queue{
		head: &seed,
		tail: &seed,
		mu:   &sync.Mutex{},
	}
	scrapedCount := 0
	jobs := make(chan string, 10)
	signal := make(chan struct{}, 1)

	// Initializing the threadpool
	for i := range 10 {
		go worker(i, jobs, signal, &scheduler, &client)
	}

	// Main scheduling logic
	for {
		scheduler.mu.Lock()
		targetURL, err := scheduler.Dequeue()
		scheduler.mu.Unlock()

		if err != nil {
			// Wait
			<-signal
			targetURL, err = scheduler.Dequeue()

			// Termination condition (But will terminate early if the first routine to send signal doesn't have any URLs in it's HTML
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
