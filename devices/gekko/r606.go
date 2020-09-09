package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/gousb"
)

type R606 struct {
	base.IDriver
}

func NewR606() *R606 {
	return &R606{
		IDriver: base.NewDriver(0x6015, 0x0403, "GekkoScience",
			"R606 Bitcoin Miner", 1, 2),
	}
}

func (r606 *R606) NewController(
	context *base.Context,
	_ base.IDriver,
	device *gousb.Device,
	inEndpoint,
	outEndpoint int,
) base.IController {
	return NewBM1387Controller(
		r606.IDriver.NewController(context, r606, device, inEndpoint, outEndpoint),
		200, 1200, 900, 12,
	)
}
