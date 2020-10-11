package base

import (
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"sync"
)

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
	for _, ct := range c.controllers {
		ct.UpdateWork(work.Clone())
	}
}
