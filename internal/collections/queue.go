package collections

import (
	"fmt"
	"sync"
)

type node struct {
	val  string
	next *node
}

type Queue struct {
	head  *node
	tail  *node
	size  int
	total int
	mu    *sync.Mutex
}

// Implement the Queue
func (q *Queue) Enqueue(item string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.size > 30000 {
		// Max queue size reached, don't add more
		return
	}
	newNode := node{
		val:  item,
		next: nil,
	}
	q.size++
	q.total++

	if q.head == nil {
		q.head = &newNode
		q.tail = &newNode
	} else {
		q.tail.next = &newNode
		q.tail = q.tail.next
	}
}

func (q *Queue) Dequeue() (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.head == nil {
		return "", fmt.Errorf("Queue is EMPTY!")
	}
	top := q.head.val
	q.head = q.head.next
	if q.head == nil {
		q.tail = nil
	}
	q.size--
	return top, nil

}

func (q *Queue) Empty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.head == nil {
		return true
	}
	return false
}

func (q *Queue) TotalQueued() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.total
}

func NewQueue() Queue {
	return Queue{
		head: nil,
		tail: nil,
		mu:   &sync.Mutex{},
	}
}
