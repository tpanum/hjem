package hjem

import "net/http"

var DefaultClient http.Client

func init() {
	DefaultClient = http.Client{
		Transport: &DefaultHeadersTripper{
			next: http.DefaultTransport,
			headers: map[string]string{
				"User-Agent": "tpanum/hjem (github.com/tpanum/hjem)",
			},
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
