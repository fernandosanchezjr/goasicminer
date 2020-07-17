package base

import "github.com/google/gousb"

type PidVid struct {
	Product gousb.ID
	Vendor  gousb.ID
}

func (p *PidVid) MatchesPidVid(desc *gousb.DeviceDesc) bool {
	return p.Product == desc.Product && p.Vendor == desc.Vendor
}
