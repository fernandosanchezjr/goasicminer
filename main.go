package main

import (
	"flag"
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/governor"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"runtime/trace"
)

var cpuProfile bool
var tracing bool
var exitChannel chan os.Signal

func init() {
	flag.BoolVar(&cpuProfile, "cpu-profile", cpuProfile, "enable cpu profiling")
	flag.BoolVar(&tracing, "trace", tracing, "enable tracing")
	exitChannel = make(chan os.Signal, 1)
}

func wait() {
	signal.Notify(exitChannel, os.Interrupt)
	signal.Notify(exitChannel, os.Kill)
	select {
	case <-exitChannel:
		return
	}
}

func main() {
	flag.Parse()
	if cpuProfile {
		f, err := os.Create("goasicminer.prof")
		if err != nil {
			panic(err)
		}
		if err = pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	}
	if tracing {
		f, err := os.Create("goasicminer.trace")
		if err != nil {
			panic(err)
		}
		if err := trace.Start(f); err != nil {
			panic(err)
		}
		defer trace.Stop()
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	gov := governor.NewGovernor(cfg)
	gov.Start()
	wait()
	gov.Stop()
}
