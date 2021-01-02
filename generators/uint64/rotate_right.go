package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type RotateRight struct {
	id     uint64
	rng    *rand.Rand
	seeded bool
}

func NewRotateRight() *RotateRight {
	return &RotateRight{
		id:     NextId(),
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
	}
}

func (rr *RotateRight) Next(previousState uint64) uint64 {
	return bits.RotateLeft64(previousState, -rr.rng.Intn(64))
}

func (rr *RotateRight) Reseed() {
	if rr.seeded {
		return
	}
	rr.seeded = true
	rr.rng.Seed(utils.RandomInt64())
}

func (rr *RotateRight) Clone() Generator64 {
	return &RotateRight{
		id:     rr.id,
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
	}
}

func (rr *RotateRight) Id() uint64 {
	return rr.id
}
