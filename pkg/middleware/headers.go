package middleware

import (
	"net/http"
)

func HeadersMiddleware(lb string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("X-Forwarded-For", r.RemoteAddr)
			r.Header.Set("X-Original-Host", r.Host)
			r.Header.Set("X-Load-Balancer", lb)

			w.Header().Set("X-Powered-By", "Goobernetes")
			w.Header().Set("X-Load-Balancer-Version", "1.0")

			next(w, r)
		}
	}
}
