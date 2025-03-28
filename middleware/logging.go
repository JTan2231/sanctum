package middleware

import (
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", time.Now().Format("2006-01-02 15:04:05"), r.Method, r.URL.Path)

		next(w, r)
	}
}
