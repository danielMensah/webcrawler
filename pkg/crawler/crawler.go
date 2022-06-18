package crawler

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/danielMensah/monzo-challenge-crawler/internal/httpclient"
	log "github.com/sirupsen/logrus"
)

// Crawler handles the crawling process
type Crawler struct {
	// baseURL is the URL to start crawling from
	baseURL string
	// httpClient is the HTTP client to use for requests
	httpClient *http.Client
	// crawledLinks is a map of links that have been crawled
	crawledLinks CrawledLinks
	// tasks is the channel to send links to be crawled
	tasks chan string
	// taskDone is the channel to receive when a task is done
	taskDone chan bool
	// wg is the wait group to wait for all tasks to be done
	wg *sync.WaitGroup
	// workers is the number of workers to use
	workers int
}

// New creates a new Crawler
func New(baseURL string, retryMax int, retryMaxWait time.Duration, workers int) *Crawler {
	c := &Crawler{
		baseURL:      baseURL,
		httpClient:   httpclient.New(retryMax, retryMaxWait),
		crawledLinks: make(map[string]bool),
		tasks:        make(chan string),
		taskDone:     make(chan bool),
		wg:           &sync.WaitGroup{},
		workers:      workers,
	}

	return c
}

// Crawl starts the crawl process
func (c *Crawler) Crawl() {
	start := time.Now()

	go c.run()
	c.tasks <- c.baseURL

	c.wg.Wait()
	close(c.taskDone)

	duration := time.Since(start)
	log.Info("Crawling finished in ", duration)
	log.Info("Found ", len(c.crawledLinks), " links")
}

// run starts the crawl process
func (c *Crawler) run() {
	for {
		select {
		case <-c.taskDone:
			return
		case link, ok := <-c.tasks:
			if !ok {
				return
			}

			log.Info("Extracting content from ", link)
			c.wg.Add(1)

			go c.extractContent(link)
		}
	}
}

// extractContent extracts the links from the given webpage
func (c *Crawler) extractContent(link string) {
	defer c.wg.Done()

	// creating custom logger to help with debugging if an error occurs during the request
	logger := log.WithField("link", link)

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		logger.WithError(err).Error("failed to create request")
		return
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.WithError(err).Error("failed to get response")
		return
	}
	defer resp.Body.Close()

	statusOK := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest
	if !statusOK {
		logger.WithField("status", resp.StatusCode).Error("request failed")
		return
	}

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.WithError(err).Error("failed to create document")
		return
	}

	logger.Info("Extracted content from ", link)
	document.Find("a").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok {
			formattedURL := formatURL(link, href)
			c.processLink(formattedURL)
		}
	})
}

// processLink checks if the link has been crawled and adds it to the tasks to be crawled
func (c *Crawler) processLink(crawledLink string) {
	if !c.crawledLinks[crawledLink] && strings.HasPrefix(crawledLink, c.baseURL) {
		c.crawledLinks[crawledLink] = true
		c.tasks <- crawledLink
	}
}

// formatURL formats the URL to be absolute
func formatURL(base string, l string) string {
	parsedUrl, err := url.Parse(l)
	if err != nil {
		return ""
	}

	if parsedUrl.IsAbs() {
		return parsedUrl.String()
	}

	parsedBase, err := url.Parse(base)
	if err != nil {
		return ""
	}

	return parsedBase.ResolveReference(parsedUrl).String()
}