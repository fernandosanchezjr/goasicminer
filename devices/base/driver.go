package base

import (
	"fmt"
	"github.com/ziutek/ftdi"
)

type IDriver interface {
	GetPidVid() PidVid
	MatchesDevice(manufacturer, productName string) bool
	String() string
	NewController(context *Context, driver IDriver, device *ftdi.Device, serialNumber string) IController
	Equals(driver IDriver) bool
	GetChannel() ftdi.Channel
}

type Driver struct {
	PidVid
	Manufacturer string
	ProductName  string
	Channel      ftdi.Channel
}

func NewDriver(product, vendor int, manufacturer, productName string, channel ftdi.Channel) *Driver {
	return &Driver{
		PidVid:       PidVid{Product: product, Vendor: vendor},
		Manufacturer: manufacturer,
		ProductName:  productName,
		Channel:      channel,
	}
}

func (d *Driver) GetPidVid() PidVid {
	return d.PidVid
}

func (d *Driver) MatchesDevice(manufaturer, productName string) bool {
	return d.Manufacturer == manufaturer && d.ProductName == productName
}

func (d *Driver) String() string {
	return fmt.Sprintf("%s %s", d.Manufacturer, d.ProductName)
}

func (d *Driver) Equals(driver IDriver) bool {
	return d.String() == driver.String()
}

func (d *Driver) NewController(
	context *Context,
	driver IDriver,
	device *ftdi.Device,
	serialNumber string,
) IController {
	return NewController(context, driver, device, serialNumber)
}

func (d *Driver) GetChannel() ftdi.Channel {
	return d.Channel
}
