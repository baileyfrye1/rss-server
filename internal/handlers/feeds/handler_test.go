package feeds

import (
	"io"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"
)

func BenchmarkProcessFeed(b *testing.B) {
	f := &Feeds{
		logger: log.New(io.Discard, "", 0),
		Client: http.Client{Timeout: 15 * time.Second},
		Cache:  &sync.Map{},
	}
	urls := []string{"https://feeds.feedburner.com/inmannews", "https://www.redfin.com/blog/feed/"}

	// Cache hit test
	for _, url := range urls {
		b.Run("cache_hit", func(b *testing.B) {
			f.Cache.Store(url, cacheEntry{
				feed:    SingularFeed{Source: "Test Feed", Posts: []Item{}},
				expires: time.Now().Add(1 * time.Hour),
			})
			for i := 0; i < b.N; i++ {
				ch := make(chan SingularFeed, 1)
				wg := &sync.WaitGroup{}
				sem := make(chan int, 1) // Serial for accurate timing
				wg.Add(1)
				sem <- 1
				go f.processFeed(url, ch, wg, sem, f.Client)
				<-ch
			}
		})
	}

	// Cache miss test
	for _, url := range urls {
		b.Run("cache_miss", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f.Cache.Delete(url)
				ch := make(chan SingularFeed, 1)
				wg := &sync.WaitGroup{}
				sem := make(chan int, 1) // Serial for accurate timing
				wg.Add(1)
				sem <- 1
				go f.processFeed(url, ch, wg, sem, f.Client)
				<-ch
			}
		})
	}
}
