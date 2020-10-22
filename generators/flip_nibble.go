package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type FlipNibble struct {
	nibble uint64
	rng    *rand.Rand
}

func NewFlipNibble(value byte) *FlipNibble {
	return &FlipNibble{
		nibble: uint64(value & 0xf),
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
	}
}

func (fn *FlipNibble) Next(previousState uint64) uint64 {
	var mask = fn.nibble << uint64(fn.rng.Intn(60))
	return previousState ^ mask
}

func (fn *FlipNibble) Reseed() {
	fn.rng.Seed(utils.RandomInt64())
}
