package ntime

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type NTime struct {
	backtrack  bool
	rng        *rand.Rand
	space      *NTimeSpace
	usedNtimes map[int]bool
}

func NewNtime() *NTime {
	n := &NTime{
		rng:        rand.New(rand.NewSource(utils.RandomInt64())),
		space:      NewNTimeSpace(),
		usedNtimes: map[int]bool{},
	}
	return n
}

func (n *NTime) Next() int {
	var next int
	var found bool
	for {
		next = n.space.offsets[n.rng.Intn(n.space.count)]
		if _, found = n.usedNtimes[next]; !found {
			n.usedNtimes[next] = true
			return next
		}
		if n.space.count == len(n.usedNtimes) {
			n.ResetUsedNtimes()
		}
	}
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

func (n *NTime) ResetUsedNtimes() {
	n.usedNtimes = map[int]bool{}
}
