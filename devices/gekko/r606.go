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
		base.NewDriver("GekkoScience", "R606 Bitcoin Miner", ftdi.ChannelA),
	}
}

func (r *R606) NewController(device *ftdi.USBDev) base.IController {
	return NewR606Controller(r.IDriver.NewController(device))
}
