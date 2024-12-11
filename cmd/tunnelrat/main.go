package main

import (
	"log"

	"git.mills.io/prologic/go-gopher"
	"github.com/tunnelchaos/tunnelrat/pkg/eventphoneSearch"
)

func main() {
	gopher.HandleFunc("/eventphone/", eventphoneSearch.Handler)
	log.Println("Listening on :7070")
	log.Fatal(gopher.ListenAndServe(":7070", nil))
}
