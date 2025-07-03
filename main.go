package main

import (
	"log"
	"net/http"
	"os"

	"github.com/baileyfrye1/rss-server/internal/handlers/feeds"
	"github.com/baileyfrye1/rss-server/internal/router"
)

func main() {
	logger := log.New(os.Stdout, "rss-server ", log.LstdFlags)
	urls := []string{"https://feeds.feedburner.com/inmannews", "https://www.redfin.com/blog/feed/"}

	fh := feeds.NewFeedHandler(logger, urls)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router.NewRouter(fh),
	}

	logger.Println("Starting server at port 8080...")

	err := server.ListenAndServe()
	if err != nil {
		logger.Fatal(err)
	}
}
