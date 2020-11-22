package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type ShiftedConstant struct {
	maskByte byte
	value    uint64
	rng      *rand.Rand
	seeded   bool
	used     int
}

func NewShiftedConstant(value byte) *ShiftedConstant {
	c := &ShiftedConstant{
		maskByte: value,
		rng:      rand.New(rand.NewSource(utils.RandomInt64())),
		seeded:   true,
	}
	c.ShuffleMask()
	return c
}

func (c *ShiftedConstant) ShuffleMask() {
	if c.maskByte == 0 {
		return
	}
	var v = uint64(c.maskByte)
	c.value = 0
	var nibbles [16]int
	for i := 0; i < 16; i++ {
		nibbles[i] = i
	}
	c.rng.Shuffle(16, func(i, j int) {
		nibbles[i], nibbles[j] = nibbles[j], nibbles[i]
	})
	var maxIndex = c.rng.Intn(16)
	for i := 0; i < maxIndex; i++ {
		c.value = c.value | v<<(nibbles[i]*4)
	}
	if c.value == 0 {
		c.value = uint64(c.maskByte)
	}
}

func (c *ShiftedConstant) Next(uint64) uint64 {
	if c.value == 0 {
		return 0
	}
	c.seeded = false
	if c.used >= MaxGeneratorReuse {
		c.ShuffleMask()
	}
	c.used += 1
	return bits.RotateLeft64(c.value, c.rng.Intn(64))
}

func (c *ShiftedConstant) Reseed() {
	if c.seeded {
		return
	}
	c.seeded = true
	c.rng.Seed(utils.RandomInt64())
}
