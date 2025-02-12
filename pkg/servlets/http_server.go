package servlets

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/raydatray/goobernetes/pkg/middleware"
	"github.com/raydatray/goobernetes/pkg/router"
)

type HttpServer struct {
	router router.RequestRouter
	port   int
	server *http.Server
}

func NewHttpServer(router router.RequestRouter, port int) *HttpServer {
	return &HttpServer{
		router: router,
		port:   port,
	}
}

func (s *HttpServer) Start() error {
	mux := http.NewServeMux()

	handler := middleware.Chain(
		middleware.HeadersMiddleware(fmt.Sprintf("goobernetes-lb-%d", s.port)),
	)(s.router.ServeRequest)

	mux.HandleFunc("/", handler)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	fmt.Printf("Load balancer started on port %d\n", s.port)
	return s.server.ListenAndServe()
}

func (s *HttpServer) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}
