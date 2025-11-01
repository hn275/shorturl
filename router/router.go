package router

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

type Router struct {
	*mux.Router
	rateLimiter map[string]*session
	rwMtx       *sync.Mutex
}

func New() *Router {
	r := &Router{
		Router:      mux.NewRouter(),
		rateLimiter: make(map[string]*session),
		rwMtx:       &sync.Mutex{},
	}

	r.Use(r.rateLimitMiddleware)

	go r.rateLimit()

	return r
}

func (r *Router) rateLimit() {
	t := time.NewTicker(time.Minute)

	for {
		<-t.C
		fmt.Println("rate limit ticker event!")

		now := time.Now()

		r.rwMtx.Lock()

		for sym := range r.rateLimiter {
			s := r.rateLimiter[sym]
			fmt.Println(sym, s.lastRead)

			if s.lastRead.Add(time.Minute).Before(now) {
				delete(r.rateLimiter, sym)
				fmt.Println("expired, deleting session:", sym)
			}
		}

		r.rwMtx.Unlock()
	}
}

func (m *Router) rateLimitMiddleware(next http.Handler) http.Handler {
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
		m.rateLimiter[r.RemoteAddr] = rl
		m.rwMtx.Unlock()

		tokenChan := make(chan struct{}, 1)

		go func() {
			rl.limiter.Take()
			tokenChan <- struct{}{}
		}()

		time.Sleep(100 * time.Microsecond)

		select {
		case <-tokenChan:
			next.ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusTooManyRequests)
		}
	})
}
