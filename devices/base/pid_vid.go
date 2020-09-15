package base

type PidVid struct {
	Product int
	Vendor  int
}

//func (p *PidVid) MatchesPidVid(desc *gousb.DeviceDesc) bool {
//	return p.Product == desc.Product && p.Vendor == desc.Vendor
//}
