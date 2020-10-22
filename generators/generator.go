package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

type Generator64 interface {
	Next(previousState uint64) uint64
	Reseed()
}

type Uint64 struct {
	rng             *rand.Rand
	generators      []Generator64
	generatorsCount int
	previousState   uint64
}

func NewUint64() *Uint64 {
	u := &Uint64{
		rng: rand.New(rand.NewSource(utils.RandomInt64())),
	}
	u.previousState = u.rng.Uint64()
	return u
}

func (u *Uint64) Add(generator ...Generator64) {
	u.generators = append(u.generators, generator...)
	u.generatorsCount = len(u.generators)
}

func (u *Uint64) Next() uint64 {
	if u.generatorsCount == 0 {
		return 0
	}
	nextState := u.generators[u.rng.Intn(u.generatorsCount)].Next(u.previousState)
	if nextState != u.previousState {
		u.previousState = nextState
	} else {
		u.previousState = u.rng.Uint64()
	}
	return u.previousState
}

func (u *Uint64) Reseed() {
	u.rng.Seed(utils.RandomInt64())
	for _, g := range u.generators {
		g.Reseed()
	}
}

func NewUint64Generator() *Uint64 {
	u := NewUint64()
	for i := 0; i < 0x10; i++ {
		b := byte(i)
		u.Add(
			NewFlipBit(), NewFlipBit(), NewFlipBit(), NewFlipBit(),
			NewFlipBit(), NewFlipBit(), NewFlipBit(), NewFlipBit(),
			NewFlipNibble(b), NewFlipNibble(b), NewFlipNibble(b), NewFlipNibble(b),
			NewFlipNibble(b), NewFlipNibble(b), NewFlipNibble(b), NewFlipNibble(b),
			NewFlipByte(), NewFlipByte(),
			NewShiftedConstant(b), NewRandom(b), NewRandomMask(b),
			NewRotateLeft(), NewRotateRight(),
			&Reverse{}, &ReverseBytes{}, &Reverse{}, &ReverseBytes{},
		)
	}
	return u
}
