package http

import "net/http"

type HeaderOption func(req *http.Request)

func WithHeader(key string, value string) func(req *http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func WithCompression() HeaderOption {
	return func(req *http.Request) {
		req.Header.Set("Accept-Encoding", "gzip,deflate,br")
	}
}
