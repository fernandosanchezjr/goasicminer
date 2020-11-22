package base

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/ziutek/ftdi"
)

type IDriverCatalog interface {
	String() string
	FindControllers(config *config.Config, context *Context) ([]IController, error)
}

type DriverCatalog struct {
	Name    string
	Drivers map[PidVid][]IDriver
}

func NewDriverCatalog(name string, driver ...IDriver) *DriverCatalog {
	dc := &DriverCatalog{Name: name, Drivers: map[PidVid][]IDriver{}}
	for _, d := range driver {
		dc.addDriver(d)
	}
	return dc
}

func (dc *DriverCatalog) addDriver(driver IDriver) {
	pidVid := driver.GetPidVid()
	drivers := dc.Drivers[pidVid]
	dc.Drivers[pidVid] = append(drivers, driver)
}

func (dc *DriverCatalog) FindControllers(config *config.Config, context *Context) ([]IController, error) {
	var controllers []IController
	var devices []*ftdi.USBDev
	var ftdiDevice *ftdi.Device
	var err error
	for pidVid, drivers := range dc.Drivers {
		devices, err = ftdi.FindAll(pidVid.Vendor, pidVid.Product)
		if err != nil {
			return nil, err
		}
		for _, dev := range devices {
			for _, drv := range drivers {
				if drv.MatchesDevice(dev.Manufacturer, dev.Description) && !context.InUse(dev.Serial) {
					if ftdiDevice, err = ftdi.OpenUSBDev(dev, drv.GetChannel()); err != nil {
						return nil, err
					}
					ctrl := drv.NewController(config, context, drv, ftdiDevice, dev.Serial)
					context.Register(ctrl)
					controllers = append(controllers, ctrl)
				}
			}
		}
	}
	return controllers, nil
}

func (dc *DriverCatalog) String() string {
	return dc.Name
}
