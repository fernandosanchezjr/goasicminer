package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type RandomSubtract struct {
	id     uint64
	rng    *rand.Rand
	start  uint64
	offset uint64
	seeded bool
}

func NewRandomSubtract(offset uint64) *RandomSubtract {
	r := &RandomSubtract{
		id:     NextId(),
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
		offset: offset,
	}
	r.start = r.rng.Uint64()
	return r
}

func (r *RandomSubtract) Next(uint64) uint64 {
	r.start += r.offset
	return r.start
}

func (r *RandomSubtract) Reseed() {
	if r.seeded {
		return
	}
	r.seeded = true
	r.rng.Seed(utils.RandomInt64())
	r.start = r.rng.Uint64()
}

func (r *RandomSubtract) Clone() Generator64 {
	return &RandomSubtract{
		id:     r.id,
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
	}
}

func (r *RandomSubtract) Id() uint64 {
	return r.id
}
