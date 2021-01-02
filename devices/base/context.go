package base

import (
	"github.com/fernandosanchezjr/goasicminer/generators/ntime"
	uint642 "github.com/fernandosanchezjr/goasicminer/generators/uint64"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sync"
)

type Context struct {
	controllersMtx   sync.Mutex
	controllers      map[string]IController
	hashRates        map[string]utils.HashRate
	versionSources   map[string]*utils.VersionSource
	ntimeSpaces      map[string]*ntime.NTimeSpace
	extraNonceSpaces map[string]*uint642.Uint64
	ntimeSpace       *ntime.NTimeSpace
	extraNonceSpace  *uint642.Uint64
	lastVersionId    uint64
	hashRateChanged  bool
	rng              *rand.Rand
	workSteps        int
}

func NewContext() *Context {
	c := &Context{
		controllers:      map[string]IController{},
		hashRates:        map[string]utils.HashRate{},
		versionSources:   map[string]*utils.VersionSource{},
		ntimeSpaces:      map[string]*ntime.NTimeSpace{},
		extraNonceSpaces: map[string]*uint642.Uint64{},
		ntimeSpace:       ntime.NewNTimeSpace(),
		extraNonceSpace:  uint642.NewUint64Generator(),
		rng:              rand.New(rand.NewSource(utils.RandomInt64())),
	}
	return c
}

func (c *Context) InUse(serial string) bool {
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	_, found := c.controllers[serial]
	return found
}

func (c *Context) Register(controller IController) {
	serialNumber := controller.String()
	if c.InUse(serialNumber) {
		return
	}
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	c.controllers[serialNumber] = controller
	c.hashRateChanged = true
}

func (c *Context) Unregister(controller IController) {
	serialNumber := controller.String()
	if !c.InUse(serialNumber) {
		return
	}
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	delete(c.controllers, serialNumber)
	delete(c.hashRates, serialNumber)
	delete(c.versionSources, serialNumber)
	delete(c.ntimeSpaces, serialNumber)
	delete(c.extraNonceSpaces, serialNumber)
	c.hashRateChanged = true
}

func (c *Context) Close() {
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	for _, ct := range c.controllers {
		ct.Close()
	}
	c.controllers = map[string]IController{}
}

func (c *Context) GetControllers(driver IDriver) []IController {
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	var found []IController
	for _, ct := range c.controllers {
		if ct.Driver().Equals(driver) {
			found = append(found, ct)
		}
	}
	return found
}

func (c *Context) UpdateWork(work *stratum.Work) {
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	var totalHashRate = c.totalHashRate()
	var serialNumber string
	var deviceHashRate utils.HashRate
	var hashPower float64
	var versionSource = work.VersionsSource
	var deviceVersionSource *utils.VersionSource
	var deviceNTimeSpace *ntime.NTimeSpace
	var extraNonceSource *uint642.Uint64
	var versionChanged = c.lastVersionId != versionSource.Id
	var workClone *stratum.Work
	if c.workSteps >= 2 || c.hashRateChanged {
		c.versionSources = map[string]*utils.VersionSource{}
		c.ntimeSpaces = map[string]*ntime.NTimeSpace{}
		c.extraNonceSpaces = map[string]*uint642.Uint64{}
		versionSource.Shuffle(c.rng)
		c.ntimeSpace.Shuffle(c.rng)
		c.extraNonceSpace.Shuffle()
		c.workSteps = 0
	} else {
		c.workSteps += 1
	}
	for _, ct := range c.controllers {
		serialNumber = ct.String()
		deviceHashRate = c.hashRates[serialNumber]
		hashPower = totalHashRate.Fraction(deviceHashRate)
		workClone = work.Clone()
		deviceVersionSource = c.versionSources[serialNumber]
		deviceNTimeSpace = c.ntimeSpaces[serialNumber]
		extraNonceSource = c.extraNonceSpaces[serialNumber]
		if c.hashRateChanged {
			log.WithFields(log.Fields{
				"serial":        serialNumber,
				"hashRate":      deviceHashRate,
				"totalHashRate": totalHashRate,
				"hashPower":     hashPower,
			}).Info("Global hash rate changed")
		}
		if c.hashRateChanged || versionChanged || deviceVersionSource == nil {
			deviceVersionSource = versionSource.Clone(hashPower)
			deviceNTimeSpace = c.ntimeSpace.Clone(hashPower)
			extraNonceSource = c.extraNonceSpace.Clone(hashPower)
			c.versionSources[serialNumber] = deviceVersionSource
			c.ntimeSpaces[serialNumber] = deviceNTimeSpace
			c.extraNonceSpaces[serialNumber] = extraNonceSource
		}
		workClone.VersionsSource = deviceVersionSource
		workClone.NTimeSpace = deviceNTimeSpace
		workClone.ExtraNonceSource = extraNonceSource
		ct.UpdateWork(workClone)
	}
	c.hashRateChanged = false
}

func (c *Context) SetHashRate(serialNumber string, hashRate utils.HashRate) {
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	c.hashRates[serialNumber] = hashRate
	c.hashRateChanged = true
	c.workSteps = 0
}

func (c *Context) totalHashRate() (total utils.HashRate) {
	for _, rate := range c.hashRates {
		total += rate
	}
	return
}
