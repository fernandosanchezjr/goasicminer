package governor

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"log"
	"sync"
)

type Governor struct {
	Config   *config.Config
	Pools    []*stratum.Pool
	PoolWork stratum.PoolWorkChan
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewGovernor(cfg *config.Config) *Governor {
	poolCount := len(cfg.Pools)
	if poolCount == 0 {
		poolCount = 1
	}
	return &Governor{Config: cfg, PoolWork: make(stratum.PoolWorkChan, poolCount), quit: make(chan struct{})}
}

func (g *Governor) Start() {
	g.wg.Add(1)
	go g.workReceiver()
	for _, poolCfg := range g.Config.Pools {
		newPool := stratum.NewPool(poolCfg, g.PoolWork)
		g.Pools = append(g.Pools, newPool)
		newPool.Start()
	}
}

func (g *Governor) Stop() {
	close(g.quit)
	g.wg.Wait()
	close(g.PoolWork)
	for _, pool := range g.Pools {
		pool.Stop()
	}
}

func (g *Governor) workReceiver() {
	var work *stratum.PoolWork
	for {
		select {
		case <-g.quit:
			g.wg.Done()
			return
		case work = <-g.PoolWork:
			log.Println("Received", work)
		}
	}
}
