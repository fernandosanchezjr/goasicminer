package utils

import "fmt"

type Version uint32

func (v Version) String() string {
	return fmt.Sprintf("%08x", uint32(v))
}

func (v Version) ZeroPositions() []int {
	var ret = make([]int, 0, 32)
	for pos := 0; pos < 32; pos++ {
		if (v & (1 << pos)) == 0 {
			ret = append(ret, pos)
		}
	}
	return ret
}
