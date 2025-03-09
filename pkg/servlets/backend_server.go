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
	mux := http.NewServeMux()

	mux.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
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
