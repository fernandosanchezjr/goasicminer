package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type FlipBitsPositions [64]int

type FlipBits struct {
	id     uint64
	bits   int
	mask   uint64
	rng    *rand.Rand
	seeded bool
}

func NewFlipBits(bits int) *FlipBits {
	fb := &FlipBits{
		id:     NextId(),
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
	var bitPositions FlipBitsPositions
	for i := 0; i < 64; i++ {
		bitPositions[i] = i
	}
	fbo.rng.Shuffle(64, (&bitPositions).shuffler)
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

func (fbp *FlipBitsPositions) shuffler(i, j int) {
	fbp[i], fbp[j] = fbp[j], fbp[i]
}
