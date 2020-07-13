package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"testing"
)

func TestR606Controller_Initialize(t *testing.T) {
	defer base.CleanUp()
	gekko := NewGekkoCatalog()
	if _, err := gekko.FindDevices(); err != nil {
		t.Fatalf("%s catalog failed to find devices: %v", gekko, err)
	}
	r606 := NewR606()
	devices := base.GetDriverDevices(r606)
	if len(devices) == 0 {
		t.Skipf("No %s devices found", gekko)
	}
	for _, dev := range devices {
		if err := dev.Initialize(); err != nil {
			t.Fatal(err)
		}
		if err := dev.Reset(); err != nil {
			t.Fatal(err)
		}
	}
}
