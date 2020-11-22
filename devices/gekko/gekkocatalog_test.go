package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"testing"
)

func TestGekkoCatalog_FindDevices(t *testing.T) {
	context := base.NewContext()
	defer context.Close()
	cfg := &config.Config{}
	gekko := NewGekkoCatalog()
	if controllers, err := gekko.FindControllers(cfg, context); err == nil {
		for _, c := range controllers {
			t.Log(gekko, "catalog found device:", c.LongString())
		}
	} else {
		t.Fatalf("%s catalog error finding controllers: %v", gekko, err)
	}
	if controllers, err := gekko.FindControllers(cfg, context); err == nil {
		if len(controllers) > 0 {
			t.Fatal("Found already opened controllers")
		}
	} else {
		t.Fatalf("%s catalog error finding controllers: %v", gekko, err)
	}
}
