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

func (u *Uint64) Shuffle() {
	u.rng.Shuffle(len(u.generators), func(i, j int) {
		u.generators[i], u.generators[j] = u.generators[j], u.generators[i]
	})
}

func (u *Uint64) Reseed() {
	u.rng.Seed(utils.RandomInt64())
	u.Shuffle()
	for _, g := range u.generators {
		g.Reseed()
	}
}

func NewUint64Generator() *Uint64 {
	u := NewUint64()
	for i := 0; i < 64; i++ {
		u.Add(NewMaskFlippers(i, u.rng)...)
	}
	for i := 0; i < 0x10; i++ {
		b := byte(i)
		for j := 0; j < 4; j++ {
			u.Add(
				&Reverse{}, &ReverseBytes{}, NewRotateLeft(), NewRotateRight(),
				NewShiftedConstant(b), NewRandom(b), NewRandomMask(b),
				NewShiftedConstant(b), NewRandom(b), NewRandomMask(b),
				NewShiftedConstant(b), NewRandom(b), NewRandomMask(b),
				NewShiftedConstant(b), NewRandom(b), NewRandomMask(b),
				NewFlipBit(), NewFlipNibble(), NewFlipByte(),
				NewFlipBit(), NewFlipNibble(), NewFlipByte(),
				NewFlipBit(), NewFlipNibble(), NewFlipByte(),
				NewFlipBit(), NewFlipNibble(), NewFlipByte(),
				NewFlipBit(), NewFlipNibble(), NewFlipByte(),
				NewFlipBit(), NewFlipNibble(), NewFlipByte(),
				NewFlipBit(), NewFlipNibble(), NewFlipByte(),
				NewFlipBit(), NewFlipNibble(), NewFlipByte(),
			)
		}
	}
	u.Shuffle()
	return u
}

func NewMaskFlippers(startPos int, rng *rand.Rand) []Generator64 {
	var generators []Generator64
	positions := make([]int, 64)
	for i := 0; i < 64; i++ {
		positions[i] = i
	}
	var mask uint64
	var bitPos int
	var index int
	for i := 0; i < 24; i++ {
		if i == 0 {
			bitPos = startPos
			index = bitPos
		} else {
			index = rng.Intn(len(positions))
			bitPos = positions[index]
		}
		mask = mask | (1 << bitPos)
		if index < len(positions)-1 {
			positions = append(positions[:index], positions[index+1:]...)
		} else {
			positions = positions[:index]
		}
		generators = append(generators, NewFlipMask(mask))
	}
	return generators
}
