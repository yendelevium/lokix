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

func worker(id int, jobs <-chan string, results chan<- struct{}, queue *Queue) {
	for job := range jobs {
		byteData := internal.FetchPage(job)
		keywords, links := internal.ParseHTML(byteData, "https://en.wikipedia.org")

		log.Printf("JOB %d: URL: %s DATA: %v", id, job, keywords)

		queue.mu.Lock()
		if queue.head == nil {
			for _, link := range links {
				queue.Enqueue(link)
			}
			results <- struct{}{}
		} else {
			for _, link := range links {
				queue.Enqueue(link)
			}
		}
		queue.mu.Unlock()
	}
}

func main() {
	log.Println("BYE, lokix")
	seed := node{
		val:  "https://en.wikipedia.org/wiki/Plant",
		next: nil,
	}

	scheduler := Queue{
		head: &seed,
		tail: &seed,
		mu:   &sync.Mutex{},
	}

	jobs := make(chan string, 10)
	results := make(chan struct{}, 1)
	for i := range 10 {
		go worker(i, jobs, results, &scheduler)
	}

	scrape_count := 0

	for {
		scheduler.mu.Lock()
		scrapeURL, err := scheduler.Dequeue()
		scheduler.mu.Unlock()

		if err != nil {
			// Wait
			<-results
			scrapeURL, err = scheduler.Dequeue()

			// Termination condition (But will fail if i have 2 routines, one doesn't have any URLS and sends msg through channel)
			// If the 2nd one had the program terminates early
			if err != nil {
				close(jobs)
				close(results)
				break
			}
		}
		jobs <- scrapeURL
		scrape_count++

		// This can also stop while scraping is going on so gotta think abt that
		if scrape_count == 500 {
			close(jobs)
			close(results)
			break
		}
	}
}
