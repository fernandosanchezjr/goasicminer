package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type Random struct {
	mask uint64
	rng  *rand.Rand
}

func NewRandom(mask byte) *Random {
	r := &Random{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
	if mask != 0x0 {
		var v = uint64(mask & 0xf)
		for i := 0; i < 16; i++ {
			r.mask = r.mask | v
			v = bits.RotateLeft64(v, 4)
		}
	}
	return r
}

func (r *Random) Next(uint64) uint64 {
	if r.mask == 0 {
		return r.rng.Uint64()
	} else {
		var mask uint64 = r.mask >> uint64(r.rng.Intn(56))
		return r.rng.Uint64() & bits.RotateLeft64(mask, r.rng.Intn(56))
	}
}

func (r *Random) Reseed() {
	r.rng.Seed(utils.RandomInt64())
}
