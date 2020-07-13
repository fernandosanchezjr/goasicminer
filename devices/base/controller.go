package base

import (
	"github.com/ziutek/ftdi"
	"log"
)

type IController interface {
	String() string
	Close()
	USBDevice() *ftdi.USBDev
	Equals(other IController) bool
	Reset() error
	Initialize() error
	Connection() *ftdi.Device
}

type Controller struct {
	*ftdi.USBDev
	channel    ftdi.Channel
	connection *ftdi.Device
}

func NewController(device *ftdi.USBDev, channel ftdi.Channel) *Controller {
	return &Controller{USBDev: device, channel: channel}
}

func (c *Controller) String() string {
	return c.Serial
}

func (c *Controller) Close() {
	if c.connection != nil {
		err := c.connection.Close()
		log.Printf("Error closing controller %s: %v", c, err)
		c.connection = nil
	}
	c.USBDev.Close()
}

func (c *Controller) USBDevice() *ftdi.USBDev {
	return c.USBDev
}

func (c *Controller) Equals(other IController) bool {
	oud := other.USBDevice()
	return c.Manufacturer == oud.Manufacturer && c.Description == oud.Description && c.Serial == oud.Serial
}

func (c *Controller) Initialize() error {
	if dev, err := ftdi.OpenUSBDev(c.USBDev, c.channel); err == nil {
		c.connection = dev
	} else {
		return err
	}
	return nil
}

func (c *Controller) Reset() error {
	return nil
}

func (c *Controller) Connection() *ftdi.Device {
	return c.connection
}
