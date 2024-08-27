package api

import (
	"log"

	"net/http"
)

func CorsMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS Middleware handling request: %s %s", r.Method, r.URL.Path)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			log.Printf("Preflight request handled: %s %s", r.Method, r.URL.Path)
			return
		}

		next.ServeHTTP(w, r)
	}
}
