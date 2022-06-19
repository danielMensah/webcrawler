package crawler

import (
	_ "embed"
	"fmt"
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
		baseURL      string
		workers      int
		retryMax     int
		retryMaxWait time.Duration
		expectedErr  error
	}{
		{
			name:         "can create a new crawler",
			baseURL:      "https://example.com",
			workers:      10,
			retryMax:     3,
			retryMaxWait: 3 * time.Second,
			expectedErr:  nil,
		},
		{
			name:         "cannot create a new crawler with missing base url",
			workers:      10,
			retryMax:     3,
			retryMaxWait: 3 * time.Second,
			expectedErr:  fmt.Errorf(_errInvalidBaseURL, ""),
		},
		{
			name:         "cannot create a new crawler with 0 workers",
			baseURL:      "https://example.com",
			workers:      0,
			retryMax:     3,
			retryMaxWait: 3 * time.Second,
			expectedErr:  fmt.Errorf(_errInvalidWorkers, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.baseURL, tt.retryMax, tt.retryMaxWait, tt.workers)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
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
			htmlPage:      nil,
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

			for i, l := range tt.expectedLinks {
				tt.expectedLinks[i] = server.URL + l
			}

			c := &Crawler{
				httpClient: server.Client(),
				baseURL:    server.URL,
				tasks:      make(chan string),
				workers:    10,
				taskDone:   make(chan bool),
				taskWg:     &sync.WaitGroup{},
			}

			c.Crawl()
			visitedUrls := c.GetVisitedLinks()

			assert.ElementsMatch(t, tt.expectedLinks, visitedUrls)
		})
	}
}
