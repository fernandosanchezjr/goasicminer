package base

import (
	"errors"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/generators"
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"github.com/ziutek/ftdi"
	"sync"
)

type IController interface {
	String() string
	LongString() string
	Close()
	Exit()
	Device() *ftdi.Device
	Driver() IDriver
	Equals(other IController) bool
	Reset() error
	UpdateWork(work *node.Work)
	WorkChannel() node.WorkChan
	AllocateWriteBuffer() ([]byte, error)
	Write(data []byte) (int, error)
	AllocateReadBuffer() ([]byte, error)
	Read(data []byte) (int, error)
	SetGenerator(generator chan *generators.Generated)
	GetGenerator() chan *generators.Generated
	ExtraNonceFound(extraNonce utils.Nonce64)
}

type Controller struct {
	device        *ftdi.Device
	driver        IDriver
	serialNumber  string
	workChan      node.WorkChan
	context       *Context
	open          bool
	generatorChan chan *generators.Generated
	mtx           sync.Mutex
}

func NewController(ctx *Context, driver IDriver, device *ftdi.Device, serialNumber string) *Controller {
	return &Controller{device: device, context: ctx, driver: driver, serialNumber: serialNumber,
		workChan: make(node.WorkChan, 16), open: true}
}

func (c *Controller) String() string {
	return c.serialNumber
}

func (c *Controller) LongString() string {
	return fmt.Sprintf("%s %s", c.driver, c)
}

func (c *Controller) recover() {
	if err := recover(); err != nil {
		log.WithFields(log.Fields{
			"serial": c.String(),
			"error":  err,
		}).Error("Error recovery")
	}
}

func (c *Controller) Close() {
	defer c.recover()
	if !c.open {
		return
	}
	c.open = false
	if err := c.device.Close(); err != nil {
		log.WithFields(log.Fields{
			"serial": c.String(),
			"error":  err,
		}).Warnln("Close error")
	}
}

func (c *Controller) Exit() {
	if !c.open {
		return
	}
	c.Close()
	c.context.Unregister(c)
}

func (c *Controller) Device() *ftdi.Device {
	return c.device
}

func (c *Controller) Driver() IDriver {
	return c.driver
}

func (c *Controller) Equals(other IController) bool {
	return c.String() == other.String()
}

func (c *Controller) Reset() error {
	return nil
}

func (c *Controller) UpdateWork(work *node.Work) {
	select {
	case c.workChan <- work:
	default:
	}
}

func (c *Controller) WorkChannel() node.WorkChan {
	return c.workChan
}

func (c *Controller) AllocateWriteBuffer() ([]byte, error) {
	if size, err := c.device.WriteChunkSize(); err != nil {
		return nil, err
	} else {
		return make([]byte, size), nil
	}
}

func (c *Controller) Write(data []byte) (int, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if !c.open {
		return 0, errors.New("device closed")
	}
	return c.device.Write(data)
}

func (c *Controller) AllocateReadBuffer() ([]byte, error) {
	if size, err := c.device.ReadChunkSize(); err != nil {
		return nil, err
	} else {
		return make([]byte, size), nil
	}
}

func (c *Controller) Read(data []byte) (int, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if !c.open {
		return 0, errors.New("device closed")
	}
	return c.device.Read(data)
}

func (c *Controller) SetGenerator(generator chan *generators.Generated) {
	c.generatorChan = generator
}

func (c *Controller) GetGenerator() chan *generators.Generated {
	return c.generatorChan
}

func (c *Controller) ExtraNonceFound(extraNonce utils.Nonce64) {
	c.context.ExtraNonceFound(extraNonce)
}
