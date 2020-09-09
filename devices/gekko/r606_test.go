package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/fernandosanchezjr/gousb"
	"testing"
	"time"
)

func TestR606_Search(t *testing.T) {
	context := base.NewContext()
	defer context.Close()
	r606 := NewR606()
	if devices, err := context.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return r606.MatchesPidVid(desc)
	}); err != nil {
		t.Fatal("context.OpenDevices error:", err)
	} else {
		for _, d := range devices {
			t.Log(r606, "driver found device at", d)
			if err := d.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestR606_Mine(t *testing.T) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	context := base.NewContext()
	defer context.Close()
	gekko := NewGekkoCatalog()
	if _, err := gekko.FindControllers(context); err != nil {
		t.Fatalf("%s catalog failed to find devices: %v", gekko, err)
	}
	r606 := NewR606()
	devices := context.GetControllers(r606)
	if len(devices) == 0 {
		t.Skipf("No %s devices found", gekko)
	}
	for _, dev := range devices {
		if err := dev.Reset(); err != nil {
			t.Fatal(err)
		}
		dev.UpdateWork(pw)
		time.Sleep(45 * time.Minute)
	}
}
