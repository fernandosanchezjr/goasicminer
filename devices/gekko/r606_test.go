package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/usb"
	"testing"
)

func TestR606_Matches(t *testing.T) {
	r606 := NewR606()
	if usbDevs, err := usb.FindFTDIDevices(); err == nil {
		for _, ud := range usbDevs {
			if r606.Matches(ud) {
				t.Log(r606, "driver found serial:", ud.Serial)
			}
			ud.Close()
		}
	} else {
		t.Fatalf("%s driver error finding devices: %v", r606, err)
	}
}
