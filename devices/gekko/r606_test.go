package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/gousb"
	"testing"
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
