package drivers

import (
	"fmt"
	"github.com/google/gousb"
	"sync/atomic"
)

var nextDriverId uint64 = 0

type USBDriver struct {
	Id              uint64
	Manufacturer    string
	ProductName     string
	SerialNumber    string
	Bus             int
	Address         int
	ConfigNumber    int
	InterfaceNumber int
	InterfaceAlt    int
	InEndpointId    int
	OutEndpointId   int
	Device          *gousb.Device
	Config          *gousb.Config
	Interface       *gousb.Interface
	InEndpoint      *gousb.InEndpoint
	OutEndpoint     *gousb.OutEndpoint
}

func (ud *USBDriver) Matches(manufacturer, productName string) bool {
	return ud.Manufacturer == manufacturer && ud.ProductName == productName
}

func (ud *USBDriver) Select(dev *gousb.Device) (*USBDriver, error) {
	var serialNumber string
	var err error
	if serialNumber, err = dev.SerialNumber(); err != nil {
		return nil, err
	}
	clone := *ud
	driverClone := &clone
	driverClone.Id = atomic.AddUint64(&nextDriverId, 1)
	driverClone.Bus = dev.Desc.Bus
	driverClone.Address = dev.Desc.Address
	driverClone.SerialNumber = serialNumber
	driverClone.Device = dev
	err = driverClone.Device.Reset()
	if err != nil {
		if closeErr := driverClone.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}
	err = driverClone.SelectConfig()
	if err != nil {
		if closeErr := driverClone.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}
	err = driverClone.SelectInterface()
	if err != nil {
		if closeErr := driverClone.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}
	err = driverClone.SelectEndpoints()
	if err != nil {
		if closeErr := driverClone.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}
	return driverClone, nil
}

func (ud *USBDriver) LongName() string {
	return fmt.Sprintf("%s %s %s", ud.Manufacturer, ud.ProductName, ud.SerialNumber)
}

func (ud *USBDriver) ShortName() string {
	return ud.SerialNumber
}

func (ud *USBDriver) SelectConfig() error {
	if err := ud.Device.SetAutoDetach(true); err != nil {
		return err
	}
	config, err := ud.Device.Config(ud.ConfigNumber)
	if err == nil {
		ud.Config = config
	}
	return err
}

func (ud *USBDriver) SelectInterface() error {
	if ud.Config == nil {
		return fmt.Errorf("no configuration selected for %s", ud.ShortName())
	}
	iface, err := ud.Config.Interface(ud.InterfaceNumber, ud.InterfaceAlt)
	if err == nil {
		ud.Interface = iface
	}
	return err
}

func (ud *USBDriver) SelectEndpoints() error {
	var err error
	ud.InEndpoint, err = ud.Interface.InEndpoint(ud.InEndpointId)
	if err != nil {
		return err
	}
	ud.OutEndpoint, err = ud.Interface.OutEndpoint(ud.OutEndpointId)
	return err
}

func (ud *USBDriver) Close() error {
	var err error
	ud.InEndpoint = nil
	ud.OutEndpoint = nil
	if ud.Interface != nil {
		ud.Interface.Close()
		ud.Interface = nil
	}
	if ud.Config != nil {
		err = ud.Config.Close()
		ud.Config = nil
		if err != nil {
			return err
		}
	}
	if ud.Device != nil {
		err = ud.Device.Close()
		ud.Device = nil
		if err != nil {
			return err
		}
	}
	ud.Bus = -1
	ud.Address = -1
	return nil
}

func (ud *USBDriver) GetAddress() (int, int) {
	return ud.Bus, ud.Address
}

func (ud *USBDriver) Equals(other Driver) bool {
	bus, address := other.GetAddress()
	return ud.Bus == bus && ud.Address == address
}

func (ud *USBDriver) InUse(bus, address int) bool {
	return ud.Bus == bus && ud.Address == address
}
