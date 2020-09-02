package governor

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"log"
	"os"
	"sync"
	"time"
)

type Governor struct {
	Config   *config.Config
	Context  *base.Context
	Catalogs []base.IDriverCatalog
	Pools    []*stratum.Pool
	PoolWork stratum.PoolWorkChan
	workQuit chan struct{}
	wg       sync.WaitGroup
}

func NewGovernor(cfg *config.Config) *Governor {
	poolCount := len(cfg.Pools)
	if poolCount == 0 {
		log.Println("No pools configured!")
		os.Exit(-1)
	}
	return &Governor{
		Context:  base.NewContext(),
		Catalogs: []base.IDriverCatalog{gekko.NewGekkoCatalog()},
		Config:   cfg,
		PoolWork: make(stratum.PoolWorkChan, poolCount),
		workQuit: make(chan struct{}),
	}
}

func (g *Governor) Start() {
	g.wg.Add(1)
	g.DeviceScan()
	go g.workReceiver()
	for _, poolCfg := range g.Config.Pools {
		newPool := stratum.NewPool(poolCfg, g.PoolWork)
		g.Pools = append(g.Pools, newPool)
		newPool.Start()
	}
}

func (g *Governor) Stop() {
	close(g.workQuit)
	g.wg.Wait()
	close(g.PoolWork)
	for _, pool := range g.Pools {
		pool.Stop()
	}
}

func (g *Governor) DeviceScan() {
	for _, cg := range g.Catalogs {
		if controllers, err := cg.FindControllers(g.Context); err == nil {
			for _, ct := range controllers {
				if err := ct.Reset(); err != nil {
					log.Printf("Error resetting %s: %s", ct, err)
				}
			}
		}
	}
}

func (g *Governor) workReceiver() {
	var work *stratum.Work
	deviceScanTicker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-g.workQuit:
			deviceScanTicker.Stop()
			g.Context.Close()
			g.wg.Done()
			return
		case work = <-g.PoolWork:
			log.Println("Received", work)
			g.Context.UpdateWork(work)
		case <-deviceScanTicker.C:
			g.DeviceScan()
		}
	}
}
