package servlets

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type BackendServer struct {
	port   int
	server *http.Server
}

func NewBackendServer(port int) *BackendServer {
	return &BackendServer{
		port: port,
	}
}

func (s *BackendServer) Start() error {
	s.server = &http.Server{
		Addr: fmt.Sprintf(":%d", s.port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Backend-Server", fmt.Sprintf("backend-%d", s.port))
			fmt.Fprintf(w, "Backend Server Port: %d\n\n", s.port)

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
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       90 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	fmt.Printf("Starting backend server on port %d\n", s.port)
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
