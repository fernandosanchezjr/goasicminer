package base

import (
	"flag"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"sync"
)

var UseRandomExtraNonce2 bool
var UseBiasedExtraNonce2 bool

func init() {
	flag.BoolVar(&UseRandomExtraNonce2, "use-random-extra-nonce", false, "use random ExtraNonce2")
	flag.BoolVar(&UseBiasedExtraNonce2, "use-biased-extra-nonce", false, "use biased random ExtraNonce2")
}

type Context struct {
	controllersMtx sync.Mutex
	controllers    map[string]IController
}

func NewContext() *Context {
	c := &Context{
		controllers: map[string]IController{},
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
}

func (c *Context) Unregister(controller IController) {
	serialNumber := controller.String()
	if !c.InUse(serialNumber) {
		return
	}
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	delete(c.controllers, serialNumber)
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
	if UseRandomExtraNonce2 {
		utils.SeedMT()
	}
	for _, ct := range c.controllers {
		if UseRandomExtraNonce2 {
			work.SetExtraNonce2(utils.RandomUint64())
		} else if UseBiasedExtraNonce2 {
			work.SetExtraNonce2(utils.Random(8.0))
		} else {
			work.ExtraNonce2 = 0
		}
		ct.UpdateWork(work.Clone())
	}
}
