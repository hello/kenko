package main

import (
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"log"
	"os/exec"
)

type MemoryUsage struct {
	total, free, buffers, cache, swap uint64
}

func (m *MemoryUsage) Usage() (float64, error) {
	memInfo, err := linuxproc.ReadMemInfo("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	m.free = memInfo["MemFree"] + memInfo["Buffers"] + memInfo["Cached"]
	m.total = memInfo["MemTotal"]

	return float64(1 - float64(m.free)/float64(m.total)), nil
}

func (m *MemoryUsage) Ranking() string {
	out, _ := exec.Command("sh", "-c", "ps -eo pmem,pid,comm | sort -nrb -k1 | head -10").Output()

	s := string(out[:])

	return fmt.Sprint("used\n\n", s)
}

func (m *MemoryUsage) Collect(queue chan *Metric) {

	metric := NewMetric()
	val, err := m.Usage()
	if err != nil {
		log.Printf("Not sending %s metric because of: %s\n", metric.Service, err)
		return
	}

	metric.Service = "memory"
	metric.Value = val

	metric.Description = m.Ranking()

	queue <- metric
}

func NewMemoryUsage() *MemoryUsage {
	return &MemoryUsage{}
}
