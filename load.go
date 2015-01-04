package main

import (
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"log"
)

type LoadAverage struct {
	last1m, last5m, last15m float64
}

func (l *LoadAverage) Usage() {
	loadAverage, avgErr := linuxproc.ReadLoadAvg("/proc/loadavg")
	cpuInfo, cpuErr := linuxproc.ReadCPUInfo("/proc/cpuinfo")
	if avgErr != nil || cpuErr != nil {
		log.Println(avgErr)
		log.Println(cpuErr)
		return
	}

	l.last1m = loadAverage.Last1Min / float64(cpuInfo.NumCPU())
	l.last5m = loadAverage.Last5Min / float64(cpuInfo.NumCPU())
	l.last15m = loadAverage.Last15Min / float64(cpuInfo.NumCPU())
}

func (l *LoadAverage) Ranking() string {
	return fmt.Sprintf("1-minute load average/core is %f", l.last1m)
}

func (l *LoadAverage) Collect(queue chan *Metric) {

	metric := NewMetric()
	l.Usage()

	metric.Service = "load1m"
	metric.Value = l.last1m
	metric.Description = "load 1m / core"

	queue <- metric

	metric = NewMetric()

	metric.Service = "load5m"
	metric.Value = l.last5m
	metric.Description = "load 5m / core"

	queue <- metric

	metric = NewMetric()

	metric.Service = "load15m"
	metric.Value = l.last15m
	metric.Description = "load 15m / core"

	queue <- metric
}

func NewLoadAverage() *LoadAverage {
	return &LoadAverage{}
}
