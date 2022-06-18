package httpclient

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

func New(retryMax int, retryWaitMax time.Duration) *http.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = retryMax
	client.RetryWaitMax = retryWaitMax
	client.Logger = nil

	c := client.StandardClient()
	c.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return client.StandardClient()
}
