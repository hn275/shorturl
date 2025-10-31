package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/ratelimit"
)

type session struct {
	limiter  ratelimit.Limiter
	lastRead time.Time
}

type router struct {
	*mux.Router
	rateLimiter map[string]*session
	rwMtx       *sync.Mutex
}

func makeRouter() *router {
	r := &router{
		Router:      mux.NewRouter(),
		rateLimiter: make(map[string]*session),
		rwMtx:       &sync.Mutex{},
	}

	r.Use(r.rateLimitMiddleware)

	go r.rateLimit()

	return r
}

func (r *router) rateLimit() {
	t := time.NewTicker(time.Minute)

	for {
		<-t.C

		now := time.Now()

		r.rwMtx.Lock()

		for sym := range r.rateLimiter {
			s := r.rateLimiter[sym]

			if s.lastRead.Add(time.Minute).Before(now) {
				delete(r.rateLimiter, sym)
				fmt.Println("expired, deleting session:", sym)
			}
		}

		r.rwMtx.Unlock()
	}
}

func (m *router) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		m.rwMtx.Lock()
		rl, ok := m.rateLimiter[r.RemoteAddr]
		if !ok {
			rl = &session{
				limiter:  ratelimit.New(5, ratelimit.Per(10*time.Second)),
				lastRead: time.Time{},
			}
		}
		rl.lastRead = now
		m.rwMtx.Unlock()

		tooManyRequest := rl.limiter.Take().Before(now)
		if tooManyRequest {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
