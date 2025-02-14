package routing

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewProxyHandler(host string) (http.Handler, error) {
	proxyUrl, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(proxyUrl)
	return proxy, nil
}
