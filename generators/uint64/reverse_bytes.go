package uint64

import (
	"math/bits"
)

type ReverseBytes struct {
}

func (*ReverseBytes) Next(previousState uint64) uint64 {
	return bits.ReverseBytes64(previousState)
}

func (*ReverseBytes) Reseed() {
}
