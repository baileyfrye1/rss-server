package feeds

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"sync"
)

type RSS struct {
	Channel Channel `xml:"channel" json:"channel"`
}

type Channel struct {
	Title string `xml:"title" json:"rssTitle"`
	Link  string `xml:"link"  json:"rssLink"`
	Items []Item `xml:"item"  json:"posts"`
}

type Item struct {
	Title   string `xml:"title"   json:"title"`
	Link    string `xml:"link"    json:"link"`
	PubDate string `xml:"pubDate" json:"pubDate"`
	Content string `xml:"encoded" json:"encoded"`
}

type Feeds struct {
	logger *log.Logger
	Urls   []string
}

func NewFeedHandler(l *log.Logger, urls []string) *Feeds {
	return &Feeds{logger: l, Urls: urls}
}

func (f *Feeds) GetFeed(w http.ResponseWriter, r *http.Request) {
	wg := sync.WaitGroup{}
	var output []byte

	wg.Add(len(f.Urls))
	for _, url := range f.Urls {
		go func() {
			defer wg.Done()
			data, _ := fetchUrl(url, f)

			rss := RSS{}

			if err := xml.Unmarshal(data, &rss); err != nil {
				f.logger.Println(err)
				return
			}

			js, err := json.MarshalIndent(&rss, "", " ")
			if err != nil {
				f.logger.Println(err)
				return
			}

			output = append(output, js...)
		}()
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

func fetchUrl(url string, f *Feeds) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		f.logger.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
