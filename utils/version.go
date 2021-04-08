package utils

import "fmt"

type Version uint32
type Versions [4]Version

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

func (v *Versions) Len() int {
	return len(v)
}

func (v *Versions) Less(i, j int) bool {
	return uint32(v[i]) < uint32(v[j])
}

func (v *Versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
