package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"testing"
	"time"
)

func TestR606Controller_Initialize(t *testing.T) {
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
		time.Sleep(20 * time.Minute)
	}
}
