package base

import (
	"github.com/fernandosanchezjr/goasicminer/generators"
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
)

type Context struct {
	controllersMtx sync.Mutex
	controllers    map[string]IController
	lastVersionId  uint64
	rng            *rand.Rand
	generator      generators.Generator
}

func NewContext() *Context {
	c := &Context{
		controllers: map[string]IController{},
		rng:         rand.New(rand.NewSource(utils.RandomInt64())),
		generator:   generators.NewUsedNTimes(),
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
	controller.SetGenerator(c.generator.GeneratorChan())
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
	c.generator.Close()
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

func (c *Context) UpdateWork(work *node.Work) {
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	c.generator.UpdateWork(work.Clone())
	for _, ct := range c.controllers {
		ct.UpdateWork(work.Clone())
	}
}

func (c *Context) ExtraNonceFound(extraNonce utils.Nonce64) {
	c.generator.ExtraNonceFound(extraNonce)
}

func (c *Context) ProgressChan() chan utils.Nonce64 {
	return c.generator.ProgressChan()
}
