package ntime

import "math/rand"

const MaxNTimeOffset = 600
const HalfNTimeOffset = 300

type NTimeSpace struct {
	offsets []int
	count   int
	pos     int
}

func NewNTimeSpace() *NTimeSpace {
	var ret = &NTimeSpace{}
	ret.offsets = make([]int, MaxNTimeOffset)
	ret.count = MaxNTimeOffset
	for i := 0; i < MaxNTimeOffset; i++ {
		ret.offsets[i] = i - HalfNTimeOffset
	}
	return ret
}

func (n *NTimeSpace) Shuffle(rng *rand.Rand) {
	rng.Shuffle(len(n.offsets), n.shuffler)
}

func (n *NTimeSpace) shuffler(i, j int) {
	n.offsets[i], n.offsets[j] = n.offsets[j], n.offsets[i]
}

func (n *NTimeSpace) Clone(fraction float64) *NTimeSpace {
	var ret = &NTimeSpace{}
	var offsetCount = len(n.offsets)
	var fractionCount = int(float64(offsetCount) * fraction)
	ret.offsets = make([]int, fractionCount)
	for i := 0; i < fractionCount; i++ {
		if n.pos >= offsetCount {
			n.pos = 0
		}
		ret.offsets[i] = n.offsets[n.pos]
		n.pos += 1
	}
	ret.count = fractionCount
	return ret
}
