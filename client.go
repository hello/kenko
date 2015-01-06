package main

import (
	"github.com/bigdatadev/goryman"
	"log"
	"time"
)

type Metric struct {
	Service, Description, State string
	Value                       interface{}
}

func NewMetric() *Metric {
	return &Metric{State: "ok"}
}

type Threshold struct {
	Warning, Critical float64
}

func NewThreshold() *Threshold {
	return &Threshold{}
}

type Kenko struct {
	exitOnSendError bool
	EventHost       string
	Interval        int
	Tag             []string
	Ttl             float32
	Ifaces          map[string]bool
	IgnoreIfaces    map[string]bool
	Thresholds      map[string]*Threshold
	Checks          map[string]bool
	client          *goryman.GorymanClient
}

func NewKenko(address string, exitOnSendError bool) *Kenko {
	return &Kenko{
		Thresholds:      make(map[string]*Threshold),
		client:          goryman.NewGorymanClient(address),
		exitOnSendError: exitOnSendError,
	}
}

func (r *Kenko) Start(done chan bool) {
	err := r.client.Connect()
	if err != nil {
		log.Fatalf("[kenko] Can not connect to host: %s\n", err)
	}
	cputime := NewCPUTime()
	memoryusage := NewMemoryUsage()
	loadaverage := NewLoadAverage()
	diskUsage := NewDiskUsage("/dev/xvda1")

	// channel size has to be large enough
	// to allow Kenko send all metrics to Riemann
	// in g.Interval
	var collectQueue chan *Metric = make(chan *Metric, 100)

	ticker := time.NewTicker(time.Second * time.Duration(r.Interval))
	go r.Report(collectQueue, done)
	running := true
	for running {
		select {
		case <-ticker.C:
			if r.Checks["cpu"] {
				go cputime.Collect(collectQueue)
			}
			if r.Checks["memory"] {
				go memoryusage.Collect(collectQueue)
			}
			if r.Checks["load"] {
				go loadaverage.Collect(collectQueue)
			}

			if r.Checks["disk"] {
				go diskUsage.Collect(collectQueue)
			}
		case <-done:
			log.Println("[kenko] Attempting to close riemann client")
			r.client.Close()
			running = false
			break

		}
	}
}

func (r *Kenko) EnforceState(metric *Metric) {

	threshold, present := r.Thresholds[metric.Service]

	if present {
		value := metric.Value

		// TODO threshold checking
		// only for int and float type
		switch {
		case value.(float64) > threshold.Critical:
			metric.State = "critical"
		case value.(float64) > threshold.Warning:
			metric.State = "warning"
		default:
			metric.State = "ok"
		}
	}
}

func (r *Kenko) Report(reportQueue chan *Metric, done chan bool) {

	for {
		select {
		case metric := <-reportQueue:
			r.EnforceState(metric)
			err := r.client.SendEvent(&goryman.Event{
				Metric:      metric.Value,
				Ttl:         r.Ttl,
				Service:     metric.Service,
				Description: metric.Description,
				Tags:        r.Tag,
				Host:        r.EventHost,
				State:       metric.State,
			})

			if err != nil {
				log.Printf("[kenko] Could not send metrics: %s\n", err)
			}
			if r.exitOnSendError {
				log.Println("[kenko] Shutting down because exitOnSendError is set to true")
				done <- true
				break
			}
		case <-done:
			log.Println("[kenko] stopping reporters")
			close(reportQueue)
			break
		}

	}
}
