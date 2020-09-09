package gekko

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/gousb"
)

type NewPac struct {
	base.IDriver
}

func NewNewPac() *NewPac {
	return &NewPac{
		IDriver: base.NewDriver(0x6015, 0x0403, "GekkoScience",
			"NewPac Bitcoin Miner", 1, 2),
	}
}

func (np *NewPac) NewController(
	context *base.Context,
	_ base.IDriver,
	device *gousb.Device,
	inEndpoint,
	outEndpoint int,
) base.IController {
	return NewBM1387Controller(
		np.IDriver.NewController(context, np, device, inEndpoint, outEndpoint),
		100, 700, 500, 2,
	)
}
