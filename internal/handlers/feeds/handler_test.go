package feeds

import (
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

func BenchmarkFetchUrl(b *testing.B) {
	f := &Feeds{logger: log.New(io.Discard, "", 0)}
	client := http.Client{Timeout: 15 * time.Second}
	urls := []string{"https://feeds.feedburner.com/inmannews", "https://www.redfin.com/blog/feed/"}

	for _, url := range urls {
		b.Run(url, func(b *testing.B) {
			for b.Loop() {
				_, err := f.fetchUrl(client, url)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
