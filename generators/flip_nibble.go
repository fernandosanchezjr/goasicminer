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
	if previousState == 0 || previousState == 0xffffffffffffffff {
		return fn.rng.Uint64()
	}
	var startPos = uint64(fn.rng.Intn(16))
	var mask = fn.nibble << startPos * 4
	if startPos > 0 {
		return previousState ^ mask
	} else if startPos == 0 {
		return previousState ^ fn.nibble
	}
	return fn.rng.Uint64()
}
