package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/ziutek/ftdi"
	"time"
)

type R606 struct {
	base.IDriver
}

func NewR606() *R606 {
	return &R606{
		IDriver: base.NewDriver(0x6015, 0x0403, "GekkoScience",
			"R606 Bitcoin Miner", ftdi.ChannelA),
	}
}

func (r606 *R606) NewController(
	config *config.Config, context *base.Context, _ base.IDriver, device *ftdi.Device, serialNumber string,
) base.IController {
	var frequency = 700.0
	for _, cfg := range config.R606 {
		if cfg.Serial == serialNumber {
			frequency = cfg.Frequency
		}
	}
	return NewBM1387Controller(
		r606.IDriver.NewController(config, context, r606, device, serialNumber),
		200, 1200, frequency, 12, 1000*time.Millisecond,
	)
}
