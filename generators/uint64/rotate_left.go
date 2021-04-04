package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type RotateLeft struct {
	rng    *rand.Rand
	seeded bool
}

func NewRotateLeft() *RotateLeft {
	return &RotateLeft{
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
	}
}

func (rl *RotateLeft) Next(previousState uint64) uint64 {
	rl.seeded = false
	return bits.RotateLeft64(previousState, rl.rng.Intn(64))
}

func (rl *RotateLeft) Reseed() {
	if rl.seeded {
		return
	}
	rl.seeded = true
	rl.rng.Seed(utils.RandomInt64())
}
