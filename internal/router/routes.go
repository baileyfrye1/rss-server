package router

import (
	"net/http"

	"github.com/baileyfrye1/rss-server/internal/handlers/feeds"
)

func NewRouter(feedHandler *feeds.Feeds) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/posts", feedHandler.GetFeed)

	return mux
}
