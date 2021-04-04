package ntime

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type NTime struct {
	backtrack bool
	rng       *rand.Rand
	space     *NTimeSpace
	ri        *utils.RandomIndex
}

func NewNtime() *NTime {
	var space = NewNTimeSpace()
	n := &NTime{
		rng:   rand.New(rand.NewSource(utils.RandomInt64())),
		space: space,
		ri:    utils.NewRandomIndex(space.count),
	}
	return n
}

func (n *NTime) Next() utils.NTime {
	var nextPos = n.ri.Next(n.rng)
	return n.space.offsets[nextPos]
}

func (n *NTime) Reseed() {
	n.rng.Seed(utils.RandomInt64())
}

func (n *NTime) Reset() {
	n.ri.Reset()
}
