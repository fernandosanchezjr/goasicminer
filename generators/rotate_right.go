package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type RotateRight struct {
	rng *rand.Rand
}

func NewRotateRight() *RotateRight {
	return &RotateRight{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
}

func (rr *RotateRight) Next(previousState uint64) uint64 {
	return bits.RotateLeft64(previousState, rr.rng.Intn(63)*-1)
}
