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

//
//var fClient = &fasthttp.Client{
//	ReadTimeout:                   1000 * time.Millisecond,
//	WriteTimeout:                  1000 * time.Millisecond,
//	MaxIdleConnDuration:           time.Hour,
//	NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
//	DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
//	DisablePathNormalizing:        true,
//
//	Dial: (&fasthttp.TCPDialer{
//		Concurrency:      4096,
//		DNSCacheDuration: time.Hour,
//	}).Dial,
//	//
//	//TLSHandshakeTimeout:   500 * time.Millisecond,
//	//DisableCompression:    true,
//	//ForceAttemptHTTP2:     true, // Forces HTTP/2 support
//}
