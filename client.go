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

type RiemannHealth struct {
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

func NewRiemannHealth(address string, exitOnSendError bool) *RiemannHealth {
	return &RiemannHealth{
		Thresholds:      make(map[string]*Threshold),
		client:          goryman.NewGorymanClient(address),
		exitOnSendError: exitOnSendError,
	}
}

func (r *RiemannHealth) Start(done chan bool) {
	err := r.client.Connect()
	if err != nil {
		log.Fatalf("[riemann-health] Can not connect to host: %s\n", err)
	}
	cputime := NewCPUTime()
	memoryusage := NewMemoryUsage()
	loadaverage := NewLoadAverage()
	diskUsage := NewDiskUsage("/dev/xvda1")

	// channel size has to be large enough
	// to allow Goshin send all metrics to Riemann
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
			log.Println("[riemann-health] Attempting to close riemann client")
			r.client.Close()
			running = false
			break

		}
	}
}

func (r *RiemannHealth) EnforceState(metric *Metric) {

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

func (r *RiemannHealth) Report(reportQueue chan *Metric, done chan bool) {

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
				log.Printf("[riemann-health] Could not send metrics: %s\n", err)
			}
			if r.exitOnSendError {
				log.Println("[riemann-health] Shutting down because exitOnSendError is set to true")
				done <- true
				break
			}
		case <-done:
			log.Println("[riemann-health] stopping reporters")
			close(reportQueue)
			break
		}

	}
}
