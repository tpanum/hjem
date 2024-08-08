package hjem

import (
	"math"
	"net/http"
	"time"
)

var DefaultClient http.Client

func init() {
	DefaultClient = http.Client{
		Transport: &RetryRoundTripper{
			next: &DefaultHeadersTripper{
				next: http.DefaultTransport,
				headers: map[string]string{
					"User-Agent": "tpanum/hjem (github.com/tpanum/hjem)",
				},
			},
			maxRetries: 5,
		},
	}
}

type DefaultHeadersTripper struct {
	next    http.RoundTripper
	headers map[string]string
}

func (t *DefaultHeadersTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Add(k, v)
	}

	return t.next.RoundTrip(req)
}

type RetryRoundTripper struct {
	next       http.RoundTripper
	maxRetries int
}

func (r *RetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= r.maxRetries; i++ {
		resp, err = r.next.RoundTrip(req)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != 429 {
			return resp, nil
		}

		baseWait := 9 * time.Second
		backoff := time.Duration(math.Pow(2, float64(i))) * time.Second
		wait := baseWait + backoff

		time.Sleep(wait)
	}

	return resp, err
}
