package governor

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko"
	"github.com/fernandosanchezjr/goasicminer/generators"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"github.com/fernandosanchezjr/goasicminer/utils"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/stianeikeland/go-rpio/v4"
	"sync"
	"time"
)

type Governor struct {
	Config     *config.Config
	Context    *base.Context
	Catalogs   []base.IDriverCatalog
	Pools      []*stratum.Pool
	PoolWork   stratum.PoolWorkChan
	Collisions map[utils.Nonce64]*protocol.Submit
	Aggregator chan *protocol.Submit
	workQuit   chan struct{}
	wg         sync.WaitGroup
	cron       *cron.Cron
	mtx        sync.Mutex
	running    bool
}

func NewGovernor(cfg *config.Config) *Governor {
	poolCount := len(cfg.Pools)
	if poolCount == 0 {
		log.Fatal("No pools configured!")
		return nil
	}
	var governor = &Governor{
		Context:    nil,
		Catalogs:   []base.IDriverCatalog{gekko.NewGekkoCatalog()},
		Config:     cfg,
		PoolWork:   nil,
		Collisions: map[utils.Nonce64]*protocol.Submit{},
		workQuit:   nil,
		cron:       cron.New(),
	}
	governor.setupTimers()
	return governor
}

func (g *Governor) setupTimers() {
	for _, downTime := range g.Config.DownTime {
		if _, err := g.cron.AddFunc(downTime.Start, g.Stop); err != nil {
			log.Fatal(err)
		}
		if _, err := g.cron.AddFunc(downTime.End, g.Start); err != nil {
			log.Fatal(err)
		}
		log.WithFields(log.Fields{
			"start": downTime.Start,
			"end":   downTime.End,
		}).Info("Downtime registered")
	}
	if len(g.cron.Entries()) > 0 {
		g.cron.Start()
	}
}

func (g *Governor) Start() {
	g.mtx.Lock()
	defer g.mtx.Unlock()
	if g.running {
		return
	}
	g.powerOn()
	log.Infoln("Starting governor")
	poolCount := len(g.Config.Pools)
	g.PoolWork = make(stratum.PoolWorkChan, poolCount*16)
	g.Aggregator = make(chan *protocol.Submit, 1024)
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
	g.powerOff()
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
						"error":  err.Error(),
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
	var found bool
	var submit, lastSubmit *protocol.Submit
	var submitChan chan *protocol.Submit
	var nonce utils.Nonce64
	deviceScanTicker := time.NewTicker(10 * time.Second)
	g.Context = base.NewContext()
	g.DeviceScan(nil)
	var progressChan = g.Context.ProgressChan()
	for {
		select {
		case <-g.workQuit:
			deviceScanTicker.Stop()
			g.Context.Close()
			g.wg.Done()
			return
		case nonce = <-progressChan:
			if submit, found = g.Collisions[nonce]; found {
				submitChan <- submit
				delete(g.Collisions, nonce)
				log.WithFields(log.Fields{
					"difficulty":  submit.Difficulty,
					"extraNonce2": submit.ExtraNonce2,
				}).Println("Aggregated expired")
			}
		case submit = <-g.Aggregator:
			if submit.Difficulty >= 1000000000 {
				submitChan <- submit
				delete(g.Collisions, nonce)
				log.WithFields(log.Fields{
					"difficulty":  submit.Difficulty,
					"extraNonce2": submit.ExtraNonce2,
				}).Println("Aggregated expired")
				continue
			}
			if lastSubmit, found = g.Collisions[submit.ExtraNonce2]; found {
				if submit.Difficulty > lastSubmit.Difficulty {
					g.Collisions[submit.ExtraNonce2] = submit
				}
			} else {
				g.Collisions[submit.ExtraNonce2] = submit
			}
		case work = <-g.PoolWork:
			g.Collisions = map[utils.Nonce64]*protocol.Submit{}
			if generators.ReuseExtraNonce2 {
				submitChan = work.SubmitChan
				work.SubmitChan = g.Aggregator
			}
			g.Context.UpdateWork(work)
		case <-deviceScanTicker.C:
			g.DeviceScan(work)
		}
	}
}

func (g *Governor) Restart() {
	g.Stop()
	g.Start()
}

func (g *Governor) Update(cfg *config.Config) {
	if len(g.cron.Entries()) > 0 {
		g.cron.Stop()
	}
	g.Config = cfg
	g.cron = cron.New()
	g.setupTimers()
	g.Restart()
}

func (g *Governor) powerOn() {
	if !g.Config.PowerControl.Enabled {
		return
	}
	if err := rpio.Open(); err != nil {
		log.WithError(err).Fatal("Could not open GPIO")
	}
	var pin = rpio.Pin(g.Config.PowerControl.Pin)
	pin.Output()
	if g.Config.PowerControl.High {
		pin.High()
	} else {
		pin.Low()
	}
}

func (g *Governor) powerOff() {
	if !g.Config.PowerControl.Enabled {
		return
	}
	var pin = rpio.Pin(g.Config.PowerControl.Pin)
	if g.Config.PowerControl.High {
		pin.Low()
	} else {
		pin.High()
	}
	if err := rpio.Close(); err != nil {
		log.WithError(err).Fatal("Could not open GPIO")
	}
	time.Sleep(10 * time.Second)
}
