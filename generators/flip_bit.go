package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type FlipBit struct {
	rng *rand.Rand
}

func NewFlipBit() *FlipBit {
	return &FlipBit{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
}

func (fbo *FlipBit) Next(previousState uint64) uint64 {
	var mask uint64 = 1 << uint64(fbo.rng.Intn(64))
	return previousState ^ mask
}
