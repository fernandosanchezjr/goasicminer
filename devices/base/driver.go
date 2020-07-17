package base

import (
	"fmt"
	"github.com/google/gousb"
)

type IDriver interface {
	MatchesPidVid(desc *gousb.DeviceDesc) bool
	MatchesDevice(manufacturer, productName string) bool
	String() string
	NewController(driver IDriver, device *gousb.Device, inEndpoint, outEndpoint int) IController
	Equals(driver IDriver) bool
	EndpointNumbers() (int, int)
}

type Driver struct {
	PidVid
	Manufacturer      string
	ProductName       string
	InEndpointNumber  int
	OutEndpointNumber int
}

func NewDriver(product, vendor gousb.ID, manufacturer, productName string, inEndpointNumber, outEndpointNumber int) *Driver {
	return &Driver{
		PidVid:            PidVid{Product: product, Vendor: vendor},
		Manufacturer:      manufacturer,
		ProductName:       productName,
		InEndpointNumber:  inEndpointNumber,
		OutEndpointNumber: outEndpointNumber,
	}
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

func (d *Driver) NewController(driver IDriver, device *gousb.Device, inEndpoint, outEndpoint int) IController {
	return NewController(driver, device, inEndpoint, outEndpoint)
}

func (d *Driver) EndpointNumbers() (int, int) {
	return d.InEndpointNumber, d.OutEndpointNumber
}
