package utils

func SwapUint64(u uint64) uint64 {
	return (u&0x00000000000000ff)<<56 |
		(u&0x000000000000ff00)<<40 |
		(u&0x0000000000ff0000)<<24 |
		(u&0x00000000ff000000)<<8 |
		(u&0x000000ff00000000)>>8 |
		(u&0x0000ff0000000000)>>24 |
		(u&0x00ff000000000000)>>40 |
		(u&0xff00000000000000)>>56
}