package router

import (
	"log/slog"
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

	go r.cacheFlush()

	return r
}

func (r *Router) cacheFlush() {
	t := time.NewTicker(time.Minute)

	for {
		<-t.C
		slog.Debug("rate limiter cache flush tick")

		now := time.Now()

		r.rwMtx.Lock()

		for sym := range r.rateLimiter {
			s := r.rateLimiter[sym]
			if s.lastRead.Add(time.Minute).Before(now) {
				delete(r.rateLimiter, sym)
				slog.Debug("removed expiring session", "addr", sym, "last-read", s.lastRead)
			}
		}

		r.rwMtx.Unlock()
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	responseCode    int
	responseBodyLen int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.responseCode = code
}

func (w *responseWriterWrapper) Write(d []byte) (int, error) {
	var err error
	w.responseBodyLen, err = w.ResponseWriter.Write(d)
	return w.responseBodyLen, err
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
