package router

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	server, err := r.lb.NextServer(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer server.ReleaseConnection()

	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s", server.GetHostPort()),
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	req.URL.Host = targetURL.Host
	req.URL.Scheme = targetURL.Scheme

	proxy.ServeHTTP(w, req)
}
