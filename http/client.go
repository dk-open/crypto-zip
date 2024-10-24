package http

import (
	"crypto/tls"
	"golang.org/x/net/http2"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var jar, _ = cookiejar.New(nil)

var http2Transport = &http2.Transport{
	TLSClientConfig: &tls.Config{
		NextProtos: []string{"h2"}, // ALPN protocol negotiation for HTTP/2
	},
}

var client = &http.Client{
	Transport: http2Transport,
	Jar:       jar,
	Timeout:   3 * time.Second,
}
