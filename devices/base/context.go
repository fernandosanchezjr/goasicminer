package base

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/fernandosanchezjr/gousb"
	"log"
	"sync"
)

type Context struct {
	*gousb.Context
	controllersMtx sync.Mutex
	controllers    []IController
}

func NewContext() *Context {
	c := &Context{
		Context: gousb.NewContext(),
	}
	c.Debug(0)
	return c
}

func (c *Context) InUse(controller IController) bool {
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	for _, ct := range c.controllers {
		if ct.Equals(controller) {
			return true
		}
	}
	return false
}

func (c *Context) Register(controller IController) {
	if c.InUse(controller) {
		return
	}
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	c.controllers = append(c.controllers, controller)
}

func (c *Context) Unregister(controller IController) {
	if !c.InUse(controller) {
		return
	}
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	var newControllers []IController
	for _, ct := range c.controllers {
		if !ct.Equals(controller) {
			newControllers = append(newControllers, ct)
		}
	}
	c.controllers = newControllers
}

func (c *Context) Close() {
	c.controllersMtx.Lock()
	defer c.controllersMtx.Unlock()
	for _, ct := range c.controllers {
		ct.Close()
	}
	c.controllers = []IController{}
	if err := c.Context.Close(); err != nil {
		log.Println("Error closing USB context:", err)
	}
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
	var randomBytes [8]byte
	for _, ct := range c.controllers {
		_, _ = rand.Read(randomBytes[:])
		work.ExtraNonce2 = binary.LittleEndian.Uint64(randomBytes[:])
		ct.UpdateWork(work.Clone())
	}
}
