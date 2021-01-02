package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type RotateLeft struct {
	id     uint64
	rng    *rand.Rand
	seeded bool
}

func NewRotateLeft() *RotateLeft {
	return &RotateLeft{
		id:     NextId(),
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

func (rl *RotateLeft) Clone() Generator64 {
	return &RotateLeft{
		id:     rl.id,
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
	}
}

func (rl *RotateLeft) Id() uint64 {
	return rl.id
}
