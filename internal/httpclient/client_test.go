package httpclient

import (
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates a standard HTTP client with retry and Scribe logging",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(3, 3*time.Second)
			assert.IsType(t, &retryablehttp.RoundTripper{}, got.Transport)

			rt := got.Transport.(*retryablehttp.RoundTripper)
			assert.Equal(t, 3, rt.Client.RetryMax)
			assert.Equal(t, float64(3), rt.Client.RetryWaitMax.Seconds())
		})
	}
}
