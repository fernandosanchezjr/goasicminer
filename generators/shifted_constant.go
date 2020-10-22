package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
	"math/rand"
)

type ShiftedConstant struct {
	value uint64
	rng   *rand.Rand
}

func NewShiftedConstant(value byte) *ShiftedConstant {
	c := &ShiftedConstant{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
	if value != 0x0 {
		var v = uint64(value & 0xf)
		for i := 0; i < 16; i++ {
			c.value = c.value | v
			v = bits.RotateLeft64(v, 4)
		}
	}
	return c
}

func (c *ShiftedConstant) Next(uint64) uint64 {
	if c.value == 0 {
		return c.value
	}
	return c.value >> c.rng.Intn(56)
}

func (c *ShiftedConstant) Reseed() {
	c.rng.Seed(utils.RandomInt64())
}
