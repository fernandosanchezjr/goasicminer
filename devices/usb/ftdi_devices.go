package usb

import (
	"github.com/ziutek/ftdi"
)

type VidPid struct {
	Vendor  int
	Product int
}

var FTDIDevices = []VidPid{
	{0x0403, 0x6015},
}

func FindFTDIDevices() ([]*ftdi.USBDev, error) {
	var devices []*ftdi.USBDev
	for _, pv := range FTDIDevices {
		if devs, err := ftdi.FindAll(pv.Vendor, pv.Product); err == nil {
			devices = append(devices, devs...)
		} else {
			return nil, err
		}
	}
	return devices, nil
}
