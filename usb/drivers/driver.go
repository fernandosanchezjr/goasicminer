package drivers

import (
	"github.com/google/gousb"
)

type Driver interface {
	Matches(manufacturer, productName string) bool
	LongName() string
	ShortName() string
	Select(dev *gousb.Device) (Driver, error)
	Initialize() error
	Close() error
	GetAddress() (int, int)
	Equals(other Driver) bool
	InUse(bus, address int) bool
}
