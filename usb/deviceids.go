package usb

import (
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/usb/drivers"
	"github.com/google/gousb"
)

type Product struct {
	ProductId gousb.ID
	Drivers   []drivers.Driver
}

type Products []Product

func (ps Products) Contains(pid gousb.ID) bool {
	for _, p := range ps {
		if p.ProductId == pid {
			return true
		}
	}
	return false
}

func (ps Products) Drivers(pid gousb.ID) []drivers.Driver {
	for _, p := range ps {
		if p.ProductId == pid {
			return p.Drivers
		}
	}
	return nil
}

type DeviceMap map[gousb.ID]Products

func (dm DeviceMap) Contains(vid, pid gousb.ID) bool {
	if productIds, found := dm[vid]; found {
		return productIds.Contains(pid)
	}
	return false
}

func (dm DeviceMap) Drivers(vid, pid gousb.ID) []drivers.Driver {
	if productIds, found := dm[vid]; found {
		return productIds.Drivers(pid)
	}
	return nil
}

func (dm DeviceMap) FindDriver(dev *gousb.Device) (drivers.Driver, error) {
	var manufacturer, productName string
	var err error
	if manufacturer, err = dev.Manufacturer(); err != nil {
		return nil, err
	}
	if productName, err = dev.Product(); err != nil {
		return nil, err
	}
	foundDrivers := dm.Drivers(dev.Desc.Vendor, dev.Desc.Product)
	for _, d := range foundDrivers {
		if d.Matches(manufacturer, productName) {
			if match, err := d.Select(dev); err == nil {
				return match, nil
			} else {
				return nil, err
			}
		}
	}
	return nil, fmt.Errorf("no matching driver found")
}

var KnownDevices = DeviceMap{
	0x0403: {
		{
			ProductId: 0x6015,
			Drivers:   []drivers.Driver{drivers.NewGekkoR606Driver()},
		},
	},
}
