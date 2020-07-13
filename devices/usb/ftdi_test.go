package usb

import (
	"testing"
)

func Test_FindFTDIDevices(t *testing.T) {
	if devices, err := FindFTDIDevices(); err == nil {
		for _, dev := range devices {
			t.Log("Found FTDI device for", dev.Manufacturer, dev.Description, dev.Serial)
			dev.Close()
		}
	} else {
		t.Fatalf("Error retrieving FTDI devices: %v", err)
	}
}
