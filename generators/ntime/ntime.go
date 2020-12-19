package ntime

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

const MaxNTimeOffset = 1024
const MinNTimeOffset = 512

type NTime struct {
	maxOffset  int
	halfOffset int
	backtrack  bool
	center     utils.NTime
	rng        *rand.Rand
}

func NewNtime() *NTime {
	n := &NTime{
		maxOffset:  MaxNTimeOffset,
		halfOffset: MaxNTimeOffset / 2,
		center:     0,
		rng:        rand.New(rand.NewSource(utils.RandomInt64())),
	}
	return n
}

func (n *NTime) Next() utils.NTime {
	var ntimeOffset = n.rng.Intn(n.maxOffset) - n.halfOffset
	return n.center + utils.NTime(ntimeOffset)
}

func (n *NTime) Reset(center utils.NTime) {
	n.center = center
}

func (n *NTime) Reseed() {
	n.rng.Seed(utils.RandomInt64())
}

func (n *NTime) Shuffle() {
	n.maxOffset = MaxNTimeOffset - n.rng.Intn(MaxNTimeOffset-MinNTimeOffset)
	n.halfOffset = n.rng.Intn(n.maxOffset / 2)
	if n.halfOffset == 0 {
		n.halfOffset = 1
	}
}
