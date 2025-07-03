package feeds

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
)

type Feeds struct {
	logger *log.Logger
	Urls   []string
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
	Content string `xml:"encoded" json:"encoded"`
}

type SingularFeed struct {
	Source string `json:"source"`
	Posts  []Item `json:"posts"`
}

type FeedsResponse struct {
	Feeds []SingularFeed `json:"feeds"`
}

func NewFeedHandler(l *log.Logger, urls []string) *Feeds {
	return &Feeds{logger: l, Urls: urls}
}

func (f *Feeds) GetFeed(w http.ResponseWriter, r *http.Request) {
	wg := sync.WaitGroup{}
	var response FeedsResponse

	wg.Add(len(f.Urls))
	for _, url := range f.Urls {
		go func() {
			defer wg.Done()

			rss := &RSS{}

			data, err := fetchUrl(url, f.logger)
			if err != nil {
				f.logger.Println(err)
				return
			}

			if err = xml.Unmarshal(data, rss); err != nil {
				f.logger.Println(err)
				return
			}

			cleanedFeed := stripHTMLTags(rss)

			sf := SingularFeed{Source: rss.Channel.Title, Posts: cleanedFeed.Channel.Items}

			response.Feeds = append(response.Feeds, sf)
		}()
	}

	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		f.logger.Println(err)
		return
	}
}

func fetchUrl(url string, logger *log.Logger) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		logger.Println(err)
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
