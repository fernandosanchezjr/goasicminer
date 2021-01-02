package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type UnshiftedRandomOrMask struct {
	id       uint64
	maskByte byte
	mask     uint64
	rng      *rand.Rand
	seeded   bool
}

func NewUnshiftedRandomOrMask(mask byte) *UnshiftedRandomOrMask {
	rm := &UnshiftedRandomOrMask{
		id:       NextId(),
		maskByte: mask,
		rng:      rand.New(rand.NewSource(utils.RandomInt64())),
		seeded:   true,
	}
	rm.ShuffleMask()
	return rm
}

func (r *UnshiftedRandomOrMask) ShuffleMask() {
	if r.maskByte == 0 {
		return
	}
	var v = uint64(r.maskByte)
	r.mask = 0
	var nibbles NibblePositions
	for i := 0; i < 16; i++ {
		nibbles[i] = i
	}
	r.rng.Shuffle(16, (&nibbles).shuffler)
	var maxIndex = r.rng.Intn(16)
	for i := 0; i < maxIndex; i++ {
		r.mask = r.mask | v<<(nibbles[i]*4)
	}
	if r.mask == 0 {
		r.mask = uint64(r.maskByte)
	}
}

func (r *UnshiftedRandomOrMask) Next(prevState uint64) uint64 {
	if r.mask == 0 {
		return r.rng.Uint64()
	}
	r.seeded = false
	r.ShuffleMask()
	//return r.rng.Uint64() | bits.RotateLeft64(r.mask, r.rng.Intn(64))
	return prevState | r.mask
}

func (r *UnshiftedRandomOrMask) Reseed() {
	if r.seeded {
		return
	}
	r.seeded = true
	r.rng.Seed(utils.RandomInt64())
}

func (r *UnshiftedRandomOrMask) Clone() Generator64 {
	return &UnshiftedRandomOrMask{
		id:       r.id,
		maskByte: r.maskByte,
		mask:     r.mask,
		rng:      rand.New(rand.NewSource(utils.RandomInt64())),
		seeded:   true,
	}
}

func (r *UnshiftedRandomOrMask) Id() uint64 {
	return r.id
}
