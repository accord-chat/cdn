package main

import (
	"accord-cdn/routes"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
)

var (
	limiters = make(map[string]throttled.RateLimiter)
	mutex    sync.Mutex
)

func getRateLimiter(path string) (throttled.RateLimiter, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if limiter, exists := limiters[path]; exists {
		return limiter, nil
	}

	store, err := memstore.New(65536)
	if err != nil {
		return nil, err
	}

	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(30),
		MaxBurst: 10,
	}

	limiter, err := throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		return nil, err
	}

	limiters[path] = limiter
	return limiter, nil
}

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limiter, err := getRateLimiter(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		limited, context, err := limiter.RateLimit(r.RemoteAddr, 1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if limited {
			w.Header().Add("Retry-After", time.Duration(context.RetryAfter).String())
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})
	http.HandleFunc("/upload", rateLimitMiddleware(routes.HandleUpload))

	log.Println("Listening on :8181")
	http.ListenAndServe(":8181", nil)
}
