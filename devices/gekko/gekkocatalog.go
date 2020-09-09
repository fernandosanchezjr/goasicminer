package gekko

import "github.com/fernandosanchezjr/goasicminer/devices/base"

type GekkoCatalog struct {
	*base.DriverCatalog
}

func NewGekkoCatalog() *GekkoCatalog {
	return &GekkoCatalog{
		base.NewDriverCatalog("GekkoScience", NewR606(), NewNewPac()),
	}
}
