package servlets

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/raydatray/goobernetes/pkg/utils"
)

type BackendServer struct {
	config utils.Config
	server *http.Server
}

func NewBackendServer(config utils.Config) *BackendServer {
	return &BackendServer{
		config: config,
	}
}

func (s *BackendServer) Start() error {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = s.config.ConnectionPoolSize
	t.MaxConnsPerHost = s.config.Connections
	t.MaxIdleConnsPerHost = s.config.ConnectionPoolSize
	t.IdleConnTimeout = 90 * time.Second
	s.server = &http.Server{
		Addr: fmt.Sprintf(":%d", s.config.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Backend-Server", fmt.Sprintf("backend-%d", s.config.Port))
			fmt.Fprintf(w, "Backend Server Port: %d\n\n", s.config.Port)

			fmt.Fprintf(w, "Request Headers:\n")
			headers := make([]string, 0, len(r.Header))
			for name := range r.Header {
				headers = append(headers, name)
			}
			sort.Strings(headers)

			for _, name := range headers {
				values := r.Header[name]
				fmt.Fprintf(w, "%s: %s\n", name, strings.Join(values, ", "))
			}
		}),
		Transport: t,
	}

	fmt.Printf("Starting backend server on port %d\n", s.config.Port)
	return s.server.ListenAndServe()
}

func (s *BackendServer) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}
