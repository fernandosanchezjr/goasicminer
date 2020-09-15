package base

import (
	"flag"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/ziutek/ftdi"
	"log"
)

var logDeviceTraffic bool

func init() {
	flag.BoolVar(&logDeviceTraffic, "log-device-traffic", false, "log device traffic")
}

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
	done         func()
	serialNumber string
	workChan     stratum.PoolWorkChan
	context      *Context
}

func NewController(ctx *Context, driver IDriver, device *ftdi.Device, serialNumber string) *Controller {
	return &Controller{device: device, context: ctx, driver: driver, done: nil, serialNumber: serialNumber,
		workChan: make(stratum.PoolWorkChan, 1)}
}

func (c *Controller) String() string {
	return c.serialNumber
}

func (c *Controller) LongString() string {
	return fmt.Sprintf("%s %s", c.driver, c)
}

func (c *Controller) recover() {
	if err := recover(); err != nil {
		log.Println("recovering from", err)
	}
}

func (c *Controller) Close() {
	defer c.recover()
	if c.done != nil {
		c.done()
		c.done = nil
	}
	if err := c.device.Close(); err != nil {
		log.Printf("Error closing %s: %s", c, err)
	}
}

func (c *Controller) Exit() {
	log.Println(c.LongString(), "exiting")
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
