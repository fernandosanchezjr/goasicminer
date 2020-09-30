package main

import (
	"flag"
	"github.com/fernandosanchezjr/goasicminer/backend/charting"
	"github.com/fernandosanchezjr/goasicminer/backend/services/implementation"
	"github.com/fernandosanchezjr/goasicminer/backend/storage"
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/logging"
	"github.com/fernandosanchezjr/goasicminer/networking/certs"
	"github.com/fernandosanchezjr/goasicminer/networking/server"
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
	flag.Parse()
	flag.BoolVar(&cpuProfile, "cpu-profile", cpuProfile, "enable cpu profiling")
	flag.BoolVar(&tracing, "trace", tracing, "enable tracing")
}

func main() {
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
	db, err := storage.GetDB()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to open aggregate DB")
	}
	if cfg.ServerAddress == "" {
		log.Fatal("Empty server address in config")
	}
	registry := services.NewRegistry()
	registry.AddService("Logging", implementation.NewLogging(db))
	registry.AddService("CheckIn", implementation.NewCheckIn(db))
	srv := server.NewServer(cfg.ServerAddress, registry)
	if err := srv.Start(); err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to start RPC server")
	}
	chartsAggregator := implementation.NewChartsAggregator(db)
	chartsAggregator.Start()
	if _, err := certs.GetCert("https"); err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to create HTTPS certificate")
	}
	cs := charting.NewService(db)
	if err := cs.Start(); err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to start HTTP server")
	}
	defer chartsAggregator.Stop()
	defer srv.Stop()
	log.Println("Backend started")
	utils.Wait()
}
