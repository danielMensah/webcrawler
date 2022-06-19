package crawler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/danielMensah/monzo-challenge-crawler/internal/httpclient"
	log "github.com/sirupsen/logrus"
)

var (
	_errInvalidBaseURL = "invalid base URL got: %s"
	_errInvalidWorkers = "workers must be greater than 0 got: %d"
)

// Crawler handles the crawling process
type Crawler struct {
	// baseURL is the URL to start crawling from
	baseURL string
	// httpClient is the HTTP client to use for requests
	httpClient *http.Client
	// crawledLinks is a map of links that have been crawled
	crawledLinks sync.Map
	// tasks is the channel to send links to be crawled
	tasks chan string
	// taskDone is the channel to receive when a task is done
	taskDone chan bool
	// taskWg is the wait group to wait for all tasks to be done
	taskWg *sync.WaitGroup
	// workers is the number of workers to use
	workers int
}

// New creates a new Crawler
func New(baseURL string, retryMax int, retryMaxWait time.Duration, workers int) (*Crawler, error) {
	if valid := isValidURL(baseURL); !valid {
		return nil, fmt.Errorf(_errInvalidBaseURL, baseURL)
	}

	if workers < 1 {
		return nil, fmt.Errorf(_errInvalidWorkers, workers)
	}

	c := &Crawler{
		baseURL:    baseURL,
		httpClient: httpclient.New(retryMax, retryMaxWait),
		tasks:      make(chan string),
		taskDone:   make(chan bool),
		taskWg:     &sync.WaitGroup{},
		workers:    workers,
	}

	return c, nil
}

// Crawl starts the crawl process
func (c *Crawler) Crawl() {
	log.Info("Starting crawl...")
	log.Info("Spawning ", c.workers, " workers")

	start := time.Now()

	wg := &sync.WaitGroup{}
	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go c.run(wg)
	}

	c.taskWg.Add(1)
	c.tasks <- c.baseURL

	c.taskWg.Wait()
	close(c.taskDone)
	wg.Wait()

	log.Info("Crawling finished in ", time.Since(start))
	log.Info("Found ", len(c.GetVisitedLinks()), " links")
}

// run starts the crawl process
func (c *Crawler) run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-c.taskDone:
			return
		case link, ok := <-c.tasks:
			if !ok {
				log.Info("tasks channel closed")
				return
			}

			log.Info("Extracting content from ", link)
			go c.extractContent(link)
		}
	}
}

// extractContent extracts the links from the given webpage
func (c *Crawler) extractContent(link string) {
	defer c.taskWg.Done()

	// creating custom logger to help with debugging if an error occurs during the request
	logger := log.WithField("link", link)

	resp, err := c.httpClient.Get(link)
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

	document.Find("a").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok {
			formattedURL := formatURL(link, href)
			c.processLink(formattedURL)
		}
	})
}

// processLink checks if the link has been crawled and adds it to the tasks to be crawled
func (c *Crawler) processLink(crawledLink string) {
	if _, found := c.crawledLinks.Load(crawledLink); !found && strings.HasPrefix(crawledLink, c.baseURL) {
		c.crawledLinks.Store(crawledLink, struct{}{})
		c.taskWg.Add(1)
		c.tasks <- crawledLink
	}
}

func (c *Crawler) GetVisitedLinks() []string {
	visitedLinks := make([]string, 0)
	c.crawledLinks.Range(func(key, value interface{}) bool {
		visitedLinks = append(visitedLinks, key.(string))
		return true
	})

	return visitedLinks
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

// isValidURL checks if the URL is valid
func isValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
