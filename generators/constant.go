package generators

import "math/bits"

type Constant struct {
	value uint64
}

func NewConstant(value byte) *Constant {
	c := &Constant{}
	if value != 0x0 {
		var v = uint64(value & 0xf)
		for i := 0; i < 16; i++ {
			c.value = c.value | v
			v = bits.RotateLeft64(v, 4)
		}
	}
	return c
}

func (c *Constant) Next(uint64) uint64 {
	return c.value
}
