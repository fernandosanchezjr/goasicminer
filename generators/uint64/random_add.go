package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type RandomAdd struct {
	id     uint64
	rng    *rand.Rand
	start  uint64
	offset uint64
	seeded bool
}

func NewRandomAdd(offset uint64) *RandomAdd {
	r := &RandomAdd{
		id:     NextId(),
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
		offset: offset,
	}
	r.start = r.rng.Uint64()
	return r
}

func (r *RandomAdd) Next(uint64) uint64 {
	r.start += r.offset
	return r.start
}

func (r *RandomAdd) Reseed() {
	if r.seeded {
		return
	}
	r.seeded = true
	r.rng.Seed(utils.RandomInt64())
	r.start = r.rng.Uint64()
}

func (r *RandomAdd) Clone() Generator64 {
	return &RandomAdd{
		id:     r.id,
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: true,
	}
}

func (r *RandomAdd) Id() uint64 {
	return r.id
}
