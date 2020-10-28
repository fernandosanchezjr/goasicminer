package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type FlipNibble struct {
	rng *rand.Rand
}

func NewFlipNibble() *FlipNibble {
	return &FlipNibble{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
}

func (fn *FlipNibble) Next(previousState uint64) uint64 {
	var mask uint64 = 0xf << uint64(fn.rng.Intn(60))
	return previousState ^ mask
}

func (fn *FlipNibble) Reseed() {
	fn.rng.Seed(utils.RandomInt64())
}
