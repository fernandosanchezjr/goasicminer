package drivers

import (
	"github.com/google/gousb"
)

const (
	GekkoR606MaxFrequencyMhz uint64 = 750
	GekkoR606MinFrequencyMhz uint64 = 100
)

type GekkoR606 struct {
	USBDriver
	ChipCount uint64
	Frequency uint64
}

func NewGekkoR606Driver() *GekkoR606 {
	return &GekkoR606{
		USBDriver: USBDriver{
			Manufacturer:    "GekkoScience",
			ProductName:     "R606 Bitcoin Miner",
			ConfigNumber:    1,
			InterfaceNumber: 0,
			InterfaceAlt:    0,
			InEndpointId:    0x81,
			OutEndpointId:   0x02,
		},
		ChipCount: 12,
		Frequency: 100,
	}
}

func (gr606 *GekkoR606) Select(dev *gousb.Device) (Driver, error) {
	clone := *gr606
	if usbDriverClone, err := gr606.USBDriver.Select(dev); err == nil {
		clone.USBDriver = *usbDriverClone
		driver := &clone
		RegisterDriver(driver)
		return driver, nil
	} else {
		return nil, err
	}
}

func (gr606 *GekkoR606) SetFrequency(value uint64) error {
	//commandBuffer := []byte{0x58, 0x09, 0x00, 0x0C, 0x00, 0x50, 0x02, 0x41, 0x00}
	if value < GekkoR606MinFrequencyMhz {
		value = GekkoR606MinFrequencyMhz
	}
	if value > GekkoR606MaxFrequencyMhz {
		value = GekkoR606MaxFrequencyMhz
	}
	return nil
}

func (gr606 *GekkoR606) Initialize() error {
	return nil
}
