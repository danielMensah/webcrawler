package main

import (
	"net/url"
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
	if valid := isValidURL(baseURL); !valid {
		log.Fatal("Invalid URL")
	}

	c := crawler.New(baseURL, 3, 60*time.Second, 50)
	c.Crawl()
}

// isValidURL checks if the URL is valid
func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
