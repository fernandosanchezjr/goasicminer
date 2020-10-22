package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type FlipByte struct {
	rng *rand.Rand
}

func NewFlipByte() *FlipByte {
	return &FlipByte{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
}

func (fbo *FlipByte) Next(previousState uint64) uint64 {
	var mask uint64 = 0xff << uint64(fbo.rng.Intn(56))
	return previousState ^ mask
}

func (fbo *FlipByte) Reseed() {
	fbo.rng.Seed(utils.RandomInt64())
}
