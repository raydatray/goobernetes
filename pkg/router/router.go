package router

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/raydatray/goobernetes/pkg/loadbalancer"
)

type RequestRouter interface {
	ServeRequest(w http.ResponseWriter, req *http.Request)
}

type Router struct {
	lb loadbalancer.LoadBalancer
}

var _ RequestRouter = (*Router)(nil)

func NewRouter(lb loadbalancer.LoadBalancer) RequestRouter {
	return &Router{
		lb: lb,
	}
}

func (r *Router) ServeRequest(w http.ResponseWriter, req *http.Request) {
	server, err := r.lb.NextServer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer server.ReleaseConnection()

	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", server.Host, server.Port),
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	req.URL.Host = targetURL.Host
	req.URL.Scheme = targetURL.Scheme

	proxy.ServeHTTP(w, req)
}
