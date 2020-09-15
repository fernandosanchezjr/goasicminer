package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/ziutek/ftdi"
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
	context *base.Context, driver base.IDriver, device *ftdi.Device, serialNumber string,
) base.IController {
	return NewBM1387Controller(
		r606.IDriver.NewController(context, r606, device, serialNumber),
		200, 1200, 900, 12,
	)
}
