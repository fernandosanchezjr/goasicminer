package ntime

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type NTime struct {
	backtrack bool
	rng       *rand.Rand
	space     *NTimeSpace
}

func NewNtime() *NTime {
	n := &NTime{
		rng:   rand.New(rand.NewSource(utils.RandomInt64())),
		space: NewNTimeSpace(),
	}
	return n
}

func (n *NTime) Next() int {
	return n.space.offsets[n.rng.Intn(n.space.count)]
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
