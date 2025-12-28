package internal

import (
	"fmt"
	"sync"
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
	q.mu.Lock()
	defer q.mu.Unlock()
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

func NewQueue() Queue {
	return Queue{
		head: nil,
		tail: nil,
		mu:   &sync.Mutex{},
	}
}
