package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type RandomXorMask struct {
	maskByte byte
	mask     uint64
	rng      *rand.Rand
	seeded   bool
	nibbles  [16]int
}

func NewRandomXorMask(mask byte) *RandomXorMask {
	rm := &RandomXorMask{
		maskByte: mask,
		rng:      rand.New(rand.NewSource(utils.RandomInt64())),
		seeded:   true,
	}
	for i := 0; i < 16; i++ {
		rm.nibbles[i] = i
	}
	rm.ShuffleMask()
	return rm
}

func (r *RandomXorMask) ShuffleMask() {
	if r.maskByte == 0 {
		return
	}
	var v = uint64(r.maskByte)
	r.mask = 0
	r.rng.Shuffle(16, r.shufflePositions)
	var maxIndex = r.rng.Intn(16)
	for i := 0; i < maxIndex; i++ {
		r.mask = r.mask | v<<(r.nibbles[i]*4)
	}
	if r.mask == 0 {
		r.mask = uint64(r.maskByte)
	}
}

func (r *RandomXorMask) shufflePositions(i, j int) {
	r.nibbles[i], r.nibbles[j] = r.nibbles[j], r.nibbles[i]
}

func (r *RandomXorMask) Next(uint64) uint64 {
	if r.mask == 0 {
		return r.rng.Uint64()
	}
	r.seeded = false
	r.ShuffleMask()
	return r.rng.Uint64() ^ bits.RotateLeft64(r.mask, r.rng.Intn(15)-7)
}

func (r *RandomXorMask) Reseed() {
	if r.seeded {
		return
	}
	r.seeded = true
	r.rng.Seed(utils.RandomInt64())
}
