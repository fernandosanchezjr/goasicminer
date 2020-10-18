package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type RotateLeft struct {
	rng *rand.Rand
}

func NewRotateLeft() *RotateLeft {
	return &RotateLeft{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
}

func (rl *RotateLeft) Next(previousState uint64) uint64 {
	return bits.RotateLeft64(previousState, rl.rng.Intn(63))
}
