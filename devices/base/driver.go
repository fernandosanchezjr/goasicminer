package base

import (
	"fmt"
	"github.com/ziutek/ftdi"
)

type IDriver interface {
	Matches(device *ftdi.USBDev) bool
	String() string
	NewController(device *ftdi.USBDev) IController
	NewDevice(controller IController) IDevice
	Equals(driver IDriver) bool
}

type Driver struct {
	Manufacturer string
	Description  string
	Channel      ftdi.Channel
}

func NewDriver(manufacturer, description string, channel ftdi.Channel) *Driver {
	return &Driver{
		Manufacturer: manufacturer,
		Description:  description,
		Channel:      channel,
	}
}

func (d *Driver) Matches(device *ftdi.USBDev) bool {
	return d.Manufacturer == device.Manufacturer && d.Description == device.Description
}

func (d *Driver) String() string {
	return fmt.Sprintf("%s %s", d.Manufacturer, d.Description)
}

func (d *Driver) Equals(driver IDriver) bool {
	return d.String() == driver.String()
}

func (d *Driver) NewController(device *ftdi.USBDev) IController {
	return NewController(device, d.Channel)
}

func (d *Driver) NewDevice(controller IController) IDevice {
	return NewDevice(d, controller)
}
