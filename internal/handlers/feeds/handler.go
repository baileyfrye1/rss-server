package feeds

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"
)

type Feeds struct {
	logger *log.Logger
	Urls   []string
	Client http.Client
}

type RSS struct {
	Channel Channel `xml:"channel" json:"channel"`
}

type Channel struct {
	Title string `xml:"title" json:"title"`
	Link  string `xml:"link"  json:"link"`
	Items []Item `xml:"item"  json:"posts"`
}

type Item struct {
	Title   string `xml:"title"   json:"title"`
	Link    string `xml:"link"    json:"link"`
	PubDate string `xml:"pubDate" json:"pubDate"`
	Content string `xml:"encoded" json:"content"`
}

type SingularFeed struct {
	Source string `json:"source"`
	Posts  []Item `json:"posts"`
}

type FeedsResponse struct {
	Feeds []SingularFeed `json:"feeds"`
}

func NewFeedHandler(l *log.Logger, urls []string) *Feeds {
	return &Feeds{logger: l, Urls: urls, Client: http.Client{Timeout: 15 * time.Second}}
}

func (f *Feeds) GetFeed(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var response FeedsResponse
	ch := make(chan SingularFeed, len(f.Urls))
	sem := make(chan int, 5)

	wg.Add(len(f.Urls))

	for _, url := range f.Urls {
		sem <- 1
		go f.processFeed(url, ch, &wg, sem, f.Client)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for v := range ch {
		response.Feeds = append(response.Feeds, v)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		f.logger.Println(err)
		return
	}
}

func (f *Feeds) fetchUrl(client http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		f.logger.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func stripHTMLTags(rss *RSS) *RSS {
	re := regexp.MustCompile(`<.*?>`)
	for i, r := range rss.Channel.Items {
		rss.Channel.Items[i].Content = re.ReplaceAllString(r.Content, "")
	}
	return rss
}

func (f *Feeds) processFeed(
	url string,
	ch chan<- SingularFeed,
	wg *sync.WaitGroup,
	sem chan int,
	client http.Client,
) {
	defer wg.Done()
	defer func() { <-sem }()

	rss := &RSS{}

	data, err := f.fetchUrl(client, url)
	if err != nil {
		f.logger.Println(err)
		return
	}

	if err = xml.Unmarshal(data, rss); err != nil {
		f.logger.Println(err)
		return
	}

	cleanedRss := stripHTMLTags(rss)

	ch <- SingularFeed{Source: rss.Channel.Title, Posts: cleanedRss.Channel.Items}
}
