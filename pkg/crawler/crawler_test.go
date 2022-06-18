package crawler

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	//go:embed testData/simple.html
	simpleHTML []byte
	//go:embed testData/duplicateLinks.html
	duplicateLinksHTML []byte
	//go:embed testData/externalLinks.html
	externalLinksHTML []byte
	//go:embed testData/invalidScheme.html
	invalidSchemeHTML []byte
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		retryMax     int
		retryMaxWait time.Duration
	}{
		{
			name:         "can create a new crawler without options",
			retryMax:     3,
			retryMaxWait: 3 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New("", tt.retryMax, tt.retryMaxWait, 1)
			assert.NotNil(t, got)
		})
	}
}

func TestCrawler_Crawl(t *testing.T) {
	tests := []struct {
		name          string
		htmlPage      []byte
		httpStatus    int
		expectedLinks []string
	}{
		{
			name:          "can crawl a page",
			htmlPage:      simpleHTML,
			httpStatus:    http.StatusOK,
			expectedLinks: []string{"/test", "/dan"},
		},
		{
			name:          "does not visits the same link twice",
			htmlPage:      duplicateLinksHTML,
			httpStatus:    http.StatusOK,
			expectedLinks: []string{"/test", "/dan"},
		},
		{
			name:          "does not visits external links",
			htmlPage:      externalLinksHTML,
			httpStatus:    http.StatusOK,
			expectedLinks: []string{"/test", "/dan"},
		},
		{
			name:          "does not visits sites with invalid scheme",
			htmlPage:      invalidSchemeHTML,
			httpStatus:    http.StatusOK,
			expectedLinks: []string{"/test", "/dan"},
		},
		{
			name:          "errors when the http status is not 200",
			htmlPage:      simpleHTML,
			httpStatus:    http.StatusNotFound,
			expectedLinks: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(tt.httpStatus)
				_, _ = w.Write(tt.htmlPage)
			}))
			defer server.Close()

			for i, l := range tt.expectedLinks {
				tt.expectedLinks[i] = server.URL + l
			}

			c := &Crawler{
				httpClient:   server.Client(),
				baseURL:      server.URL,
				crawledLinks: make(map[string]bool),
				tasks:        make(chan string),
				workers:      10,
				taskDone:     make(chan bool),
				wg:           &sync.WaitGroup{},
			}

			c.Crawl()
			visitedUrls := c.crawledLinks.List()

			assert.ElementsMatch(t, tt.expectedLinks, visitedUrls)
		})
	}
}
