package middleware

import (
	"log"
	"net/http"
	"time"
)

// loggingMiddleware is a custom middleware function that logs incoming HTTP requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the details of the incoming request.
		log.Printf("Incoming request: Method=%s, Path=%s, RemoteAddr=%s, Time=%s\n",
			r.Method, r.URL.Path, r.RemoteAddr, time.Now().Format(time.RFC3339))

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
