package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"testing"
	"time"
)

func TestR606Controller_Initialize(t *testing.T) {
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
		time.Sleep(10 * time.Second)
	}
}
