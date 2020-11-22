package governor

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	log "github.com/sirupsen/logrus"
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
		log.Fatal("No pools configured!")
		return nil
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
	g.DeviceScan(nil)
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

func (g *Governor) DeviceScan(work *stratum.Work) {
	for _, cg := range g.Catalogs {
		if controllers, err := cg.FindControllers(g.Config, g.Context); err == nil {
			for _, ct := range controllers {
				if err := ct.Reset(); err != nil {
					log.WithFields(log.Fields{
						"serial": ct.String(),
						"error":  err,
					}).Warnln("Error resetting controller")
				}
			}
			if len(controllers) != 0 && work != nil {
				g.Context.UpdateWork(work)
			}
		}
	}
}

func (g *Governor) workReceiver() {
	var work *stratum.Work
	deviceScanTicker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-g.workQuit:
			deviceScanTicker.Stop()
			g.Context.Close()
			g.wg.Done()
			return
		case work = <-g.PoolWork:
			g.Context.UpdateWork(work)
		case <-deviceScanTicker.C:
			g.DeviceScan(work)
		}
	}
}
