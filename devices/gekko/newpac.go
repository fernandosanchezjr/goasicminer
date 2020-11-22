package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/config"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/ziutek/ftdi"
)

type NewPac struct {
	base.IDriver
}

func NewNewPac() *NewPac {
	return &NewPac{
		IDriver: base.NewDriver(0x6015, 0x0403, "GekkoScience",
			"NewPac Bitcoin Miner", ftdi.ChannelA),
	}
}

func (np *NewPac) NewController(
	config *config.Config, context *base.Context, driver base.IDriver, device *ftdi.Device, serialNumber string,
) base.IController {
	return NewBM1387Controller(
		np.IDriver.NewController(config, context, np, device, serialNumber),
		100, 700, 550, 2,
	)
}
