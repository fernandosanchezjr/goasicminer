package base

import (
	"github.com/fernandosanchezjr/goasicminer/devices/usb"
	"github.com/ziutek/ftdi"
)

type IDriverCatalog interface {
	Match(device *ftdi.USBDev) IDriver
	String() string
	FindDevices() ([]IDevice, error)
}

type DriverCatalog struct {
	Name    string
	Drivers []IDriver
}

func NewDriverCatalog(name string, driver ...IDriver) *DriverCatalog {
	return &DriverCatalog{Name: name, Drivers: driver}
}

func (dc *DriverCatalog) Match(device *ftdi.USBDev) IDriver {
	for _, d := range dc.Drivers {
		if d.Matches(device) {
			return d
		}
	}
	return nil
}

func (dc *DriverCatalog) String() string {
	return dc.Name
}

func (dc *DriverCatalog) FindDevices() ([]IDevice, error) {
	var result []IDevice
	if usbDevs, err := usb.FindFTDIDevices(); err == nil {
		for _, ud := range usbDevs {
			if driver := dc.Match(ud); driver != nil {
				device := driver.NewDevice(driver.NewController(ud))
				if !deviceInUse(device) {
					registerDevice(device)
					result = append(result, device)
				}
			} else {
				ud.Close()
			}
		}
		return result, nil
	} else {
		return nil, err
	}
}
