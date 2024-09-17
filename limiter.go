package ratelimiter

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]*UserRequest
	window   time.Duration
	limit    int
}

type UserRequest struct {
	count       int
	lastRequest time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*UserRequest),
		window:   window,
		limit:    limit,
	}
}

func (rl *RateLimiter) AllowRequest(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if req, exists := rl.requests[ip]; exists {
		if time.Since(req.lastRequest) > rl.window {
			req.count = 1
			req.lastRequest = time.Now()
			log.Printf("Rate limit reset for IP %s", ip)
			return true
		}

		if req.count < rl.limit {
			req.count++
			req.lastRequest = time.Now()
			log.Printf("IP %s made request %d of %d", ip, req.count, rl.limit)
			return true
		}

		log.Printf("Rate limit exceeded for IP %s", ip)
		return false
	}

	rl.requests[ip] = &UserRequest{count: 1, lastRequest: time.Now()}
	log.Printf("First request from IP %s", ip)
	return true
}

func RateLimiterMiddleware(rl *RateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !rl.AllowRequest(ip) {
			log.Printf("Request blocked for IP %s due to rate limit", ip)
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		log.Printf("Request allowed for IP %s", ip)
		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ip := strings.Split(forwarded, ",")[0]
		return strings.TrimSpace(ip)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("Failed to parse IP from RemoteAddr: %v", err)
		return r.RemoteAddr
	}

	return host
}
