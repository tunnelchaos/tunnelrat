package main

import (
	"flag"
	"log"
	"strconv"

	"git.mills.io/prologic/go-gopher"
	"github.com/tunnelchaos/go-packages/config"
	"github.com/tunnelchaos/tunnelrat/pkg/chaospost"
	"github.com/tunnelchaos/tunnelrat/pkg/eventphoneSearch"
	"github.com/tunnelchaos/tunnelrat/pkg/mastodonHashtag"
)

var (
	configPath  string
	secretsPath string
)

func main() {
	flag.StringVar(&configPath, "config", "config.toml", "path to the config file")
	flag.StringVar(&secretsPath, "secrets", "secrets.toml", "path to the secrets file")
	flag.Parse()
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}
	secrets, err := config.LoadSecrets(secretsPath)
	if err != nil {
		log.Fatalf("Failed to load secrets file: %v", err)
	}
	gopher.HandleFunc("/eventphone/", eventphoneSearch.Handler)
	gopher.HandleFunc("/mastodon/", func(w gopher.ResponseWriter, r *gopher.Request) {
		mastodonHashtag.Handler(w, r, secrets)
	})
	gopher.HandleFunc("/chaospost/", func(w gopher.ResponseWriter, r *gopher.Request) {
		chaospost.Handler(w, r, secrets, conf.Server.Hostname, conf.Server.SearchPort)
	})
	port := strconv.Itoa(conf.Server.SearchPort)
	log.Println("Listening on", port)
	log.Fatal(gopher.ListenAndServe(":"+port, nil))
}
