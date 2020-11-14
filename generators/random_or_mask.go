package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type RandomOrMask struct {
	maskByte byte
	mask     uint64
	rng      *rand.Rand
	seeded   bool
	used     int
}

func NewRandomOrMask(mask byte) *RandomOrMask {
	rm := &RandomOrMask{
		maskByte: mask,
		rng:      rand.New(rand.NewSource(utils.RandomInt64())),
		seeded:   true,
	}
	rm.ShuffleMask()
	return rm
}

func (r *RandomOrMask) ShuffleMask() {
	var v = uint64(r.maskByte)
	r.mask = 0
	var nibbles [16]int
	for i := 0; i < 16; i++ {
		nibbles[i] = i
	}
	r.rng.Shuffle(16, func(i, j int) {
		nibbles[i], nibbles[j] = nibbles[j], nibbles[i]
	})
	var maxIndex = r.rng.Intn(16)
	for i := 0; i < maxIndex; i++ {
		r.mask = r.mask | v<<(nibbles[i]*4)
	}
	if r.mask == 0 {
		r.mask = uint64(r.maskByte)
	}
}

func (r *RandomOrMask) Next(uint64) uint64 {
	r.seeded = false
	if r.used >= MaxGeneratorReuse {
		r.ShuffleMask()
	}
	r.used += 1
	return r.rng.Uint64() | bits.RotateLeft64(r.mask, r.rng.Intn(63))
}

func (r *RandomOrMask) Reseed() {
	if r.seeded {
		return
	}
	r.seeded = true
	r.rng.Seed(utils.RandomInt64())
}