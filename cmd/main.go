package main

import (
	"os"
	"time"

	"github.com/danielMensah/monzo-challenge-crawler/pkg/crawler"
	log "github.com/sirupsen/logrus"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal("url argument is required")
	}

	baseURL := args[0]
	c, err := crawler.New(baseURL, 3, 60*time.Second, 50)
	if err != nil {
		log.WithError(err).Fatal("failed to create crawler")
	}

	c.Crawl()
}
