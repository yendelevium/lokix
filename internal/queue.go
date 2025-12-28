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
	Head *node
	Tail *node
	Mu   *sync.Mutex
}

// Implement the Queue
func (q *Queue) Enqueue(item string) {
	q.Mu.Lock()
	defer q.Mu.Unlock()
	newNode := node{
		val:  item,
		next: nil,
	}

	if q.Head == nil {
		q.Head = &newNode
		q.Tail = &newNode
	} else {
		q.Tail.next = &newNode
		q.Tail = q.Tail.next
	}
}

func (q *Queue) Dequeue() (string, error) {
	q.Mu.Lock()
	defer q.Mu.Unlock()

	if q.Head == nil {
		return "", fmt.Errorf("Queue is EMPTY!")
	}
	top := q.Head.val
	q.Head = q.Head.next
	if q.Head == nil {
		q.Tail = nil
	}
	return top, nil

}
