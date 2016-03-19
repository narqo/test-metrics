package main

import (
	"sync"
	"math/rand"
	"time"

	"github.com/rcrowley/go-metrics"
)

func startTailSender(w *sync.WaitGroup) {
	r := metrics.NewRegistry()
	c := metrics.GetOrRegisterCounter("tail.count", r)
	s := metrics.GetOrRegisterCounter("tail.size", r)

	ticker := time.NewTicker(time.Millisecond * 100)
	go func() {
		for _ = range ticker.C {
			c.Inc(1)
			if (rand.Float64() < 0.3) {
				s.Inc(1)
			}
		}
	}()

	go collect(r)

	time.Sleep(time.Second * 20)
	ticker.Stop()
	w.Done()
}

func startHTTPSender(w *sync.WaitGroup) {
	r := metrics.NewRegistry()
	c := metrics.GetOrRegisterCounter("http.count", r)

	ticker := time.NewTicker(time.Millisecond * 300)
	go func() {
		for _ = range ticker.C {
			c.Inc(1)
		}
	}()

	go collect(r)

	time.Sleep(time.Second * 20)
	ticker.Stop()
	w.Done()
}
