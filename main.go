package main

import (
	"flag"
	"fmt"
	config "github.com/stvp/go-toml-config"
	"log"
	"os"
	"os/signal"
)

var (
	hostname, _ = os.Hostname()
	configPath  = flag.String("c", "my_app.conf", "Path to config file. Ex: kenko -c /etc/hello/kenko.conf")
)

var (
	ec2MetaData     = config.Bool("ec2", true)
	riemannHost     = config.String("riemann.host", "127.0.0.1")
	riemannPort     = config.Int("riemann.port", 5555)
	exitOnSendError = config.Bool("riemann.exit_on_send_error", true)
	eventInterval   = config.Int("event.interval", 5)
	eventTtl        = config.Int("event.ttl", 10)
	cpuCheck        = config.Bool("cpu.check", true)
	cpuWarning      = config.Float64("cpu.warning", 0.9)
	cpuCritical     = config.Float64("cpu.critical", 0.95)
	loadCheck       = config.Bool("load.check", true)
	loadWarning     = config.Float64("load.warning", 3)
	loadCritical    = config.Float64("load.critical", 8)
	memoryCheck     = config.Bool("memory.check", true)
	memoryWarning   = config.Float64("memory.warning", 0.75)
	memoryCritical  = config.Float64("memory.critical", 0.9)
	diskCheck       = config.Bool("disk.check", true)
	diskWarning     = config.Float64("disk.warning", 0.6)
	diskCritical    = config.Float64("disk.critical", 0.9)
)

func main() {
	flag.Parse()
	log.Printf("[kenko] Configuration loaded from: %s\n", *configPath)

	if err := config.Parse(*configPath); err != nil {
		log.Fatal(err)
	}
	msg := "[kenko] Kenko is configured to report metrics to: %s on port %s at a %ds interval.\n"
	log.Printf(msg, *riemannHost, *riemannHost, *eventInterval)

	address := fmt.Sprintf("%s:%d", *riemannHost, *riemannPort)
	app := NewKenko(address, *exitOnSendError)

	app.EventHost = hostname
	app.Interval = *eventInterval

	if *ec2MetaData {
		log.Println("[kenko] Attempting to get tags from ec2 metadata")
		app.Tag = GetTags()
		log.Printf("[kenko] Got tags: %s", app.Tag)
	}

	app.Ttl = float32(*eventTtl)

	cpuThreshold := NewThreshold()
	cpuThreshold.Critical = *cpuCritical
	cpuThreshold.Warning = *cpuWarning

	app.Thresholds["cpu"] = cpuThreshold

	loadThreshold := NewThreshold()
	loadThreshold.Critical = *loadCritical
	loadThreshold.Warning = *loadWarning

	app.Thresholds["load1m"] = loadThreshold
	app.Thresholds["load5m"] = loadThreshold
	app.Thresholds["load15m"] = loadThreshold

	memoryThreshold := NewThreshold()
	memoryThreshold.Critical = *memoryCritical
	memoryThreshold.Warning = *memoryWarning

	app.Thresholds["memory"] = memoryThreshold

	diskThreshold := NewThreshold()
	diskThreshold.Critical = *diskCritical
	diskThreshold.Warning = *diskWarning

	app.Thresholds["disk"] = diskThreshold

	checks := make(map[string]bool)
	checks["cpu"] = *cpuCheck
	checks["load"] = *loadCheck
	checks["memory"] = *memoryCheck
	checks["disk"] = *diskCheck

	app.Checks = checks

	done := make(chan bool)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Printf("Got ctrl+c, shutting down.(%s)\n", sig)
			done <- true
		}
	}()

	app.Start(done)
	close(done)
}
