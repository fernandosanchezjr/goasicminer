package utils

func BMCRC(data []byte) {
	var i, idx, _count uint32
	var c1, c2 byte
	_count = uint32((len(data) - 1) * 8)
	c := [5]byte{1, 1, 1, 1, 1}
	for i = 0; i < _count; i++ {
		c1 = c[1]
		c[1] = c[0]
		if c2 = data[idx] & (0x80 >> (i % 8)); c2 > 0 {
			c[0] = c[4] ^ 1
		} else {
			c[0] = c[4] ^ 0
		}
		c[4] = c[3]
		c[3] = c[2]
		c[2] = c1 ^ c[0]

		if ((i + 1) % 8) == 0 {
			idx++
		}
	}
	data[len(data)-1] |= (c[4] * 0x10) | (c[3] * 0x08) | (c[2] * 0x04) | (c[1] * 0x02) | (c[0] * 0x01)
}
