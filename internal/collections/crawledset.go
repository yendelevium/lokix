package collections

import (
	"hash/fnv"
	"sync"
)

// The "bloomfilter" - I know how to implement a bloomfilter, but I don't want to implemet it again
// This projects focus is a web-crawler not a bloomfilter
type CrawledSet struct {
	crawledURLS map[uint64]bool
	mu          *sync.Mutex
}

func (c *CrawledSet) Add(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.crawledURLS[hashURL(url)] = true
}

func (c *CrawledSet) Contains(url string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.crawledURLS[hashURL(url)]
}

func (c *CrawledSet) Total() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.crawledURLS)
}

func hashURL(url string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(url))
	return h.Sum64()
}

func NewCrawledSet() CrawledSet {
	return CrawledSet{
		crawledURLS: map[uint64]bool{},
		mu:          &sync.Mutex{},
	}
}
