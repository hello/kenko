// +build !windows

package main

import (
	"os"
	"syscall"
)

type DiskUsage struct {
	volumePath string
	stat       *syscall.Statfs_t
}

// Returns an object holding the disk usage of the volume
// that volumePath belongs to
func NewDiskUsage(volumePath string) *DiskUsage {
	return &DiskUsage{volumePath: volumePath, stat: nil}
}

// Total free bytes on file system
func (d *DiskUsage) Free() uint64 {
	return d.stat.Bfree * uint64(d.stat.Bsize)
}

// Total available bytes on file system to an unpriveleged user
func (d *DiskUsage) Available() uint64 {
	return d.stat.Bavail * uint64(d.stat.Bsize)
}

// Total size of the file system
func (d *DiskUsage) Size() uint64 {
	return d.stat.Blocks * uint64(d.stat.Bsize)
}

// Total bytes used in file system
func (d *DiskUsage) Used() uint64 {
	return d.Size() - d.Free()
}

// Percentage of use on the file system
func (d *DiskUsage) Usage() float64 {
	return float64(d.Used()) / float64(d.Size())
}

func (d *DiskUsage) Collect(queue chan *Metric) {
	os.Chdir(d.volumePath)

	var stat syscall.Statfs_t
	wd, _ := os.Getwd()
	syscall.Statfs(wd, &stat)
	d.stat = &stat

	metric := NewMetric()

	metric.Service = "disk"
	metric.Value = d.Usage()
	metric.Description = "% free"

	queue <- metric
}
