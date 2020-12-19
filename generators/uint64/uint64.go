package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

const MaxBitsFlipped = 8
const MaxBitFlipperCount = 64

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
	u.previousState = u.generators[u.rng.Intn(u.generatorsCount)].Next(u.previousState)
	return u.previousState
}

func (u *Uint64) Shuffle() {
	u.rng.Shuffle(len(u.generators), func(i, j int) {
		u.generators[i], u.generators[j] = u.generators[j], u.generators[i]
	})
}

func (u *Uint64) Reset() {
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
	for i := 0; i < 4; i++ {
		u.Add(rotators...)
	}
	var reverses = []Generator64{&Reverse{}, &ReverseBytes{}}
	for i := 0; i < 4; i++ {
		u.Add(reverses...)
	}
	for i := 0; i < MaxBitsFlipped; i++ {
		var flipper = NewFlipBits(i + 1)
		for j := 0; j < MaxBitFlipperCount; j++ {
			u.Add(flipper)
		}
	}
	random := NewRandom()
	for i := 0; i < 16; i++ {
		b := byte(i)
		u.Add(
			random, NewShiftedConstant(b), NewRandomAndMask(b), NewRandomOrMask(b), NewRandomXorMask(b),
		)
	}
	u.Shuffle()
	return u
}
