package governor

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	cron "github.com/robfig/cron/v3"
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
	cron     *cron.Cron
	mtx      sync.Mutex
	running  bool
}

func NewGovernor(cfg *config.Config) *Governor {
	poolCount := len(cfg.Pools)
	if poolCount == 0 {
		log.Fatal("No pools configured!")
		return nil
	}
	var governor = &Governor{
		Context:  nil,
		Catalogs: []base.IDriverCatalog{gekko.NewGekkoCatalog()},
		Config:   cfg,
		PoolWork: nil,
		workQuit: nil,
		cron:     cron.New(),
	}
	governor.setupTimers()
	return governor
}

func (g *Governor) setupTimers() {
	if _, err := g.cron.AddFunc("00 17 * * *", g.Stop); err != nil {
		log.Fatal(err)
	}
	if _, err := g.cron.AddFunc("00 20 * * *", g.Start); err != nil {
		log.Fatal(err)
	}
	g.cron.Start()
}

func (g *Governor) Start() {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	if g.running {
		return
	}
	log.Infoln("Starting governor")
	poolCount := len(g.Config.Pools)
	g.PoolWork = make(stratum.PoolWorkChan, poolCount)
	g.wg.Add(1)
	g.workQuit = make(chan struct{})
	go g.workReceiver()
	for _, poolCfg := range g.Config.Pools {
		newPool := stratum.NewPool(poolCfg, g.PoolWork)
		g.Pools = append(g.Pools, newPool)
		newPool.Start()
	}
	g.running = true
}

func (g *Governor) Stop() {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	if !g.running {
		return
	}
	log.Infoln("Stopping governor")
	close(g.workQuit)
	g.wg.Wait()
	close(g.PoolWork)
	for _, pool := range g.Pools {
		pool.Stop()
	}
	g.Pools = nil
	g.running = false
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
	g.Context = base.NewContext()
	g.DeviceScan(nil)
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
