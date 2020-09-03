package main

import (
	"flag"
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/governor"
	"log"
	"os"
	"os/signal"
	"runtime"
)

var exitchannel = make(chan os.Signal, 1)

func wait() {
	signal.Notify(exitchannel, os.Interrupt)
	signal.Notify(exitchannel, os.Kill)
	select {
	case <-exitchannel:
		return
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	flag.Parse()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	gov := governor.NewGovernor(cfg)
	gov.Start()
	wait()
	gov.Stop()
}
