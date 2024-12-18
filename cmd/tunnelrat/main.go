package main

import (
	"log"

	"git.mills.io/prologic/go-gopher"
	"github.com/tunnelchaos/tunnelrat/pkg/eventphoneSearch"
	"github.com/tunnelchaos/tunnelrat/pkg/mastodonHashtag"
)

func main() {
	gopher.HandleFunc("/eventphone/", eventphoneSearch.Handler)
	gopher.HandleFunc("/mastodon/", mastodonHashtag.Handler)
	log.Println("Listening on :7070")
	log.Fatal(gopher.ListenAndServe(":7070", nil))
}
