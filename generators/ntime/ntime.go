package ntime

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type NTime struct {
	maxOffset  int
	halfOffset int
	backtrack  bool
	center     utils.NTime
	rng        *rand.Rand
	space      *NTimeSpace
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
	return n.center + utils.NTime(n.space.offsets[n.rng.Intn(n.space.count)])
}

func (n *NTime) Reset(center utils.NTime, space *NTimeSpace) {
	n.center = center
	n.space = space
}

func (n *NTime) Reseed() {
	n.rng.Seed(utils.RandomInt64())
}

func (n *NTime) Shuffle() {
	if n.space == nil {
		return
	}
	n.space.Shuffle(n.rng)
}
