package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type Random struct {
	id     uint64
	rng    *rand.Rand
	seeded bool
}

func NewRandom() *Random {
	r := &Random{
		id:     NextId(),
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
	}
	return r
}

func (r *Random) Next(uint64) uint64 {
	r.seeded = false
	return r.rng.Uint64()
}

func (r *Random) Reseed() {
	if r.seeded {
		return
	}
	r.seeded = true
	r.rng.Seed(utils.RandomInt64())
}

func (r *Random) Clone() Generator64 {
	return &Random{
		id:     r.id,
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
	}
}

func (r *Random) Id() uint64 {
	return r.id
}
