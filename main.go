package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

const (
	flushPrefix = "five_sec"
	flushInterval = time.Second * 5
	batchSize = 500
)

var (
	queue []*Datum
	qlock sync.Mutex
	dchan chan *Datum
)

func init() {
	qlock = sync.Mutex{}
	dchan = make(chan *Datum)
}

func main() {
	log.Println("start")

	go queuer()
	go send()

	var wg sync.WaitGroup
	wg.Add(2)

	go startTailSender(&wg)
	go startHTTPSender(&wg)

	wg.Wait()
	log.Println("stop")
}

func queuer() {
	for {
		select {
		case d := <-dchan:
			qlock.Lock()
			log.Printf("queueing datum %s\n", d)
			queue = append(queue, d)
			qlock.Unlock()
		default:
		}
	}
}

func send() {
	for _ = range time.Tick(time.Millisecond * 500) {
		qlock.Lock()
		if i := len(queue); i > 0 {
			if i > batchSize {
				i = batchSize
			}
			sending := queue[:i]
			queue = queue[i:]
			log.Printf("sending: %d, remaining: %d\n", i, len(queue))
			qlock.Unlock()
			sendBatch(sending)
		} else {
			qlock.Unlock()
			log.Println("nothing to send")
		}
	}
}

func sendBatch(batch []*Datum) {
	log.Println("send batch")
}

func collect(r metrics.Registry) {
	for _ = range time.Tick(flushInterval) {
		flush(r, flushPrefix)
	}
}

func flush(r metrics.Registry, prefix string) {
	now := time.Now()
	r.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			d := &Datum{
				Name:      fmt.Sprintf("%s.%s", prefix, name),
				Value:     metric.Count(),
				Timestamp: now,
			}
			log.Printf("flushing datum %s\n", d)
			dchan <- d
		}
	})
}

type Datum struct {
	Name      string
	Value     int64
	Timestamp time.Time
}

func (d Datum) String() string {
	return fmt.Sprintf("%s=%d", d.Name, d.Value)
}
