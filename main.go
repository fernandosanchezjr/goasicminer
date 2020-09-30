package main

import (
	"flag"
	client2 "github.com/fernandosanchezjr/goasicminer/backend/services/shim"
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/governor"
	"github.com/fernandosanchezjr/goasicminer/logging"
	"github.com/fernandosanchezjr/goasicminer/networking/client"
	"github.com/fernandosanchezjr/goasicminer/networking/services"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime/pprof"
	"runtime/trace"
)

var cpuProfile bool
var tracing bool

func init() {
	flag.BoolVar(&cpuProfile, "cpu-profile", cpuProfile, "enable cpu profiling")
	flag.BoolVar(&tracing, "trace", tracing, "enable tracing")
}

func main() {
	flag.Parse()
	logging.SetupLogger()
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
	if cfg.BackendAddress != "" {
		registry := services.NewRegistry()
		registry.AddService("Logging", client2.NewLogging())
		registry.AddService("CheckIn", client2.NewCheckIn())
		cl := client.NewClient(cfg.BackendAddress, registry)
		cl.Start()
		defer cl.Stop()
		logIngestHook := logging.NewIngestHook(cl)
		log.AddHook(logIngestHook)
		if _, err := cl.Call("CheckIn", "Host", logIngestHook.HostName); err != nil {
			log.WithError(err).Error("CheckIn error")
		}
	}
	gov := governor.NewGovernor(cfg)
	gov.Start()
	utils.Wait()
	gov.Stop()
}
