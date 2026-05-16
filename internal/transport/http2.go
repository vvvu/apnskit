package transport

import (
	"crypto/tls"
	"net/http"
	"time"
)

// NewHTTP2Client creates an HTTP Client which enforces HTTP/2 and TLS 1.2+
func NewHTTP2Client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   32,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
}
