package base

import (
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	log "github.com/sirupsen/logrus"
	"github.com/ziutek/ftdi"
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
	UpdateWork(work *stratum.Work)
	WorkChannel() stratum.PoolWorkChan
	AllocateWriteBuffer() ([]byte, error)
	Write(data []byte) (int, error)
	AllocateReadBuffer() ([]byte, error)
	Read(data []byte) (int, error)
}

type Controller struct {
	device       *ftdi.Device
	driver       IDriver
	serialNumber string
	workChan     stratum.PoolWorkChan
	context      *Context
	open         bool
}

func NewController(ctx *Context, driver IDriver, device *ftdi.Device, serialNumber string) *Controller {
	return &Controller{device: device, context: ctx, driver: driver, serialNumber: serialNumber,
		workChan: make(stratum.PoolWorkChan, 16), open: true}
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
		}).Warnln("Error recovery")
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
	log.WithFields(log.Fields{
		"serial": c.String(),
	}).Infoln("Controller exiting")
	c.Close()
	defer c.recover()
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

func (c *Controller) UpdateWork(work *stratum.Work) {
	select {
	case c.workChan <- work:
	default:
	}
}

func (c *Controller) WorkChannel() stratum.PoolWorkChan {
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
	return c.device.Read(data)
}
