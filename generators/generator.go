package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

const MaxBitsFlipped = 8
const MaxFlipBits = 16
const MaxGeneratorReuse = 4

type Generator64 interface {
	Next(previousState uint64) uint64
	Reseed()
}

type Uint64 struct {
	rng               *rand.Rand
	generators        []Generator64
	generatorsCount   int
	previousGenerated map[uint64]byte
	previousState     uint64
}

func NewUint64() *Uint64 {
	u := &Uint64{
		rng:               rand.New(rand.NewSource(utils.RandomInt64())),
		previousGenerated: map[uint64]byte{},
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
	var nextState uint64
	for {
		nextState = u.generators[u.rng.Intn(u.generatorsCount)].Next(u.previousState)
		if _, found := u.previousGenerated[nextState]; !found {
			u.previousState = nextState
			u.previousGenerated[nextState] = 0
			return nextState
		}
	}
}

func (u *Uint64) Shuffle() {
	u.rng.Shuffle(len(u.generators), func(i, j int) {
		u.generators[i], u.generators[j] = u.generators[j], u.generators[i]
	})
}

func (u *Uint64) Reset() {
	u.previousGenerated = map[uint64]byte{}
}

func (u *Uint64) Reseed() {
	u.rng.Seed(utils.RandomInt64())
	for _, g := range u.generators {
		g.Reseed()
	}
}

func NewUint64Generator() *Uint64 {
	u := NewUint64()
	var rotators = []Generator64{NewRotateLeft(), NewRotateRight()}
	for i := 0; i < 16; i++ {
		u.Add(rotators...)
	}
	var reversers = []Generator64{&Reverse{}, &ReverseBytes{}}
	for i := 0; i < 14; i++ {
		u.Add(reversers...)
	}
	for i := 0; i < MaxBitsFlipped; i++ {
		var flipper = NewFlipBits(i)
		for j := 0; j < MaxFlipBits; j++ {
			u.Add(flipper)
		}
	}
	var zero = &Zero{}
	for i := 0; i < 3; i++ {
		u.Add(zero)
	}
	var random = NewRandom()
	for i := 0; i < 13; i++ {
		u.Add(random)
	}
	for i := 1; i < 0x10; i++ {
		b := byte(i)
		u.Add(
			random, NewShiftedConstant(b), NewRandomAndMask(b), NewRandomOrMask(b),
		)
	}
	u.Shuffle()
	return u
}
