package httpclient

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

var defaultClient = &http.Client{
	Timeout: 6 * time.Second,
	Transport: &http.Transport{
		DisableKeepAlives: true,
		MaxIdleConns:      100,
		IdleConnTimeout:   90 * time.Second,
	},
}

func NewDefaultHttpClient() *resty.Client {
	return resty.NewWithClient(defaultClient).SetRetryCount(3)
}
