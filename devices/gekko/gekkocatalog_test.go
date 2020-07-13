package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"testing"
)

func TestGekkoCatalog_FindDevices(t *testing.T) {
	defer base.CleanUp()
	gekko := NewGekkoCatalog()
	if devices, err := gekko.FindDevices(); err == nil {
		for _, d := range devices {
			t.Log(gekko, "catalog found device:", d.LongString())
		}
	} else {
		t.Fatalf("%s catalog error finding devices: %v", gekko, err)
	}
	if devices, err := gekko.FindDevices(); err == nil {
		if len(devices) > 0 {
			t.Fatal("Found already opened devices")
		}
	} else {
		t.Fatalf("%s catalog error finding devices: %v", gekko, err)
	}
}
