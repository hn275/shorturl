package router

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

const (
	testAddr = "0.0.0.0:8000"
)

func TestRouterRateLimiter(t *testing.T) {
	r := New()

	r.Handle(
		"/healthcheck",
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	)

	go func() {
		http.ListenAndServe(testAddr, r)
	}()

	time.Sleep(time.Second)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/healthcheck", testAddr), nil)
	if err != nil {
		t.Fatal(err)
	}

	cx := &http.Client{}

	if _, err := cx.Do(req); err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	for i := range 10 {
		response, err := cx.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("request", i, response.Status)
	}

	wg := &sync.WaitGroup{}

	for i := range 5 {
		wg.Go(func() {
			cx := &http.Client{}

			var request http.Request
			request = *req

			if _, err := cx.Do(&request); err != nil {
				t.Fatal(err)
			}

			time.Sleep(10 * time.Second)

			for j := range 10 {
				response, err := cx.Do(&request)
				if err != nil {
					t.Fatal(err)
				}
				t.Log("thread", i, "request", j, response.Status)
			}
		})
	}

	wg.Wait()

	t.Log("waiting for rl map clear, 2 ticks")
	time.Sleep(2*time.Minute + time.Second)

	if len(r.rateLimiter) != 0 {
		t.Fatal("rate limiter map not cleared")
	}
}
