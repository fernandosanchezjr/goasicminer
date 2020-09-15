package base

import (
	"github.com/ziutek/ftdi"
)

type IDriverCatalog interface {
	String() string
	FindControllers(context *Context) ([]IController, error)
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

func (dc *DriverCatalog) FindControllers(context *Context) ([]IController, error) {
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
					ctrl := drv.NewController(context, drv, ftdiDevice, dev.Serial)
					context.Register(ctrl)
					controllers = append(controllers, ctrl)
				}
			}
		}
	}
	return controllers, nil
}

//func (dc *DriverCatalog) MatchesPidVid(desc *gousb.DeviceDesc) bool {
//	for _, d := range dc.Drivers {
//		if d.MatchesPidVid(desc) {
//			return true
//		}
//	}
//	return false
//}

//func (dc *DriverCatalog) MatchesDevice(device *gousb.Device) (IDriver, error) {
//	var manufacturer, productname string
//	var err error
//	if manufacturer, err = device.Manufacturer(); err != nil {
//		return nil, err
//	}
//	if productname, err = device.Product(); err != nil {
//		return nil, err
//	}
//	for _, d := range dc.Drivers {
//		if d.MatchesDevice(manufacturer, productname) {
//			return d, nil
//		}
//	}
//	return nil, nil
//}

func (dc *DriverCatalog) String() string {
	return dc.Name
}

//func (dc *DriverCatalog) FindControllers(context *Context) ([]IController, error) {
//	var result []IController
//	var devices []*gousb.Device
//	var driver IDriver
//	var err error
//	if devices, err = context.OpenDevices(func(desc *gousb.DeviceDesc) bool {
//		return dc.MatchesPidVid(desc)
//	}); err != nil {
//		return nil, err
//	}
//	for _, d := range devices {
//		if driver, err = dc.MatchesDevice(d); err != nil {
//			return nil, err
//		} else if driver != nil {
//			in, out := driver.EndpointNumbers()
//			controller := driver.NewController(context, driver, d, in, out)
//			if !context.InUse(controller) {
//				if err = controller.Initialize(); err != nil {
//					controller.Close()
//					return nil, err
//				}
//				context.Register(controller)
//				result = append(result, controller)
//			}
//		} else {
//			if err = d.Close(); err != nil {
//				return nil, err
//			}
//		}
//	}
//	return result, nil
//}
