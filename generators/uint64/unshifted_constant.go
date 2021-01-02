package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type UnshiftedConstant struct {
	id       uint64
	maskByte byte
	value    uint64
	rng      *rand.Rand
	seeded   bool
}

func NewUnshiftedConstant(value byte) *UnshiftedConstant {
	c := &UnshiftedConstant{
		id:       NextId(),
		maskByte: value,
		rng:      rand.New(rand.NewSource(utils.RandomInt64())),
		seeded:   true,
	}
	c.ShuffleMask()
	return c
}

func (c *UnshiftedConstant) ShuffleMask() {
	if c.maskByte == 0 {
		return
	}
	var v = uint64(c.maskByte)
	c.value = 0
	var nibbles NibblePositions
	for i := 0; i < 16; i++ {
		nibbles[i] = i
	}
	c.rng.Shuffle(16, (&nibbles).shuffler)
	var maxIndex = c.rng.Intn(16)
	for i := 0; i < maxIndex; i++ {
		c.value = c.value | v<<(nibbles[i]*4)
	}
	if c.value == 0 {
		c.value = uint64(c.maskByte)
	}
}

func (c *UnshiftedConstant) Next(uint64) uint64 {
	if c.value == 0 {
		return 0
	}
	c.seeded = false
	c.ShuffleMask()
	return c.value
}

func (c *UnshiftedConstant) Reseed() {
	if c.seeded {
		return
	}
	c.seeded = true
	c.rng.Seed(utils.RandomInt64())
}

func (c *UnshiftedConstant) Clone() Generator64 {
	return &UnshiftedConstant{
		id:       c.id,
		maskByte: c.maskByte,
		value:    c.value,
		rng:      rand.New(rand.NewSource(utils.RandomInt64())),
		seeded:   true,
	}
}

func (c *UnshiftedConstant) Id() uint64 {
	return c.id
}
