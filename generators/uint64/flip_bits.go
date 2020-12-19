package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type FlipBits struct {
	bits   int
	mask   uint64
	rng    *rand.Rand
	seeded bool
	used   int
}

func NewFlipBits(bits int) *FlipBits {
	fb := &FlipBits{
		bits:   bits,
		rng:    rand.New(rand.NewSource(utils.RandomInt64())),
		seeded: false,
	}
	fb.ShuffleMask()
	return fb
}

func (fbo *FlipBits) ShuffleMask() {
	if fbo.bits == 1 {
		fbo.mask = 1
		return
	}
	var bitPositions [64]int
	for i := 0; i < 64; i++ {
		bitPositions[i] = i
	}
	fbo.rng.Shuffle(64, func(i, j int) {
		bitPositions[i], bitPositions[j] = bitPositions[j], bitPositions[i]
	})
	fbo.mask = 0
	for i := 0; i < fbo.bits; i++ {
		fbo.mask = fbo.mask | 1<<bitPositions[i]
	}
}

func (fbo *FlipBits) Next(previousState uint64) uint64 {
	fbo.seeded = false
	fbo.ShuffleMask()
	return previousState ^ fbo.mask
}

func (fbo *FlipBits) Reseed() {
	if fbo.seeded {
		return
	}
	fbo.seeded = true
	fbo.rng.Seed(utils.RandomInt64())
}
