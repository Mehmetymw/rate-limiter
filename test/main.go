package main

import (
	"log"
	"net/http"
	"time"

	ratelimiter "github.com/mehmetymw/rate-limiter"
)

func main() {
	rl := ratelimiter.NewRateLimiter(5, time.Minute)
	mux := http.NewServeMux()
	mux.Handle("/", ratelimiter.RateLimiterMiddleware(rl, http.HandlerFunc(mainHandler)))

	log.Println("Server running on :9090")
	err := http.ListenAndServe(":9090", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
func mainHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Request received"))
}
