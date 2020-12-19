package uint64

import (
	"math/bits"
)

type Reverse struct {
}

func (*Reverse) Next(previousState uint64) uint64 {
	return bits.Reverse64(previousState)
}

func (*Reverse) Reseed() {
}
