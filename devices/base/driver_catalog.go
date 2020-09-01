package base

import (
	"github.com/google/gousb"
)

type IDriverCatalog interface {
	MatchesPidVid(desc *gousb.DeviceDesc) bool
	MatchesDevice(device *gousb.Device) (IDriver, error)
	String() string
	FindDevices(context *Context) ([]IController, error)
}

type DriverCatalog struct {
	Name    string
	Drivers []IDriver
}

func NewDriverCatalog(name string, driver ...IDriver) *DriverCatalog {
	return &DriverCatalog{Name: name, Drivers: driver}
}

func (dc *DriverCatalog) MatchesPidVid(desc *gousb.DeviceDesc) bool {
	for _, d := range dc.Drivers {
		if d.MatchesPidVid(desc) {
			return true
		}
	}
	return false
}

func (dc *DriverCatalog) MatchesDevice(device *gousb.Device) (IDriver, error) {
	var manufacturer, productname string
	var err error
	if manufacturer, err = device.Manufacturer(); err != nil {
		return nil, err
	}
	if productname, err = device.Product(); err != nil {
		return nil, err
	}
	for _, d := range dc.Drivers {
		if d.MatchesDevice(manufacturer, productname) {
			return d, nil
		}
	}
	return nil, nil
}

func (dc *DriverCatalog) String() string {
	return dc.Name
}

func (dc *DriverCatalog) FindControllers(context *Context) ([]IController, error) {
	var result []IController
	if devices, err := context.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return dc.MatchesPidVid(desc)
	}); err != nil {
		return nil, err
	} else {
		for _, d := range devices {
			if driver, err := dc.MatchesDevice(d); err != nil {
				return nil, err
			} else if driver != nil {
				in, out := driver.EndpointNumbers()
				controller := driver.NewController(driver, d, in, out)
				if !context.InUse(controller) {
					if err := controller.Initialize(); err != nil {
						controller.Close()
						return nil, err
					}
					context.Register(controller)
					result = append(result, controller)
				}
			} else {
				if err := d.Close(); err != nil {
					return nil, err
				}
			}
		}
	}
	return result, nil
}
