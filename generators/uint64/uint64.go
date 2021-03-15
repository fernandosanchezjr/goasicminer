package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync/atomic"
)

const MaxBitsFlipped = 16
const MaxBitFlipperCount = 4

var id uint64

type Uint64 struct {
	rng               *rand.Rand
	generators        []utils.Generator64
	generatorsCount   int
	previousState     uint64
	previousGenerated map[uint64]byte
	pos               int
}

func NewUint64() *Uint64 {
	u := &Uint64{
		rng:               rand.New(rand.NewSource(utils.RandomInt64())),
		previousGenerated: map[uint64]byte{},
	}
	u.previousState = u.rng.Uint64()
	return u
}

func (u *Uint64) Add(generator ...utils.Generator64) {
	u.generators = append(u.generators, generator...)
	u.generatorsCount = len(u.generators)
}

func (u *Uint64) Next() uint64 {
	if u.generatorsCount == 0 {
		return 0
	}
	var nextState uint64
	var found bool
	for {
		nextState = u.generators[u.rng.Intn(u.generatorsCount)].Next(u.previousState)
		if _, found = u.previousGenerated[nextState]; !found {
			u.previousState = nextState
			u.previousGenerated[nextState] = 0
			return nextState
		}
	}
}

func (u *Uint64) Shuffle() {
	u.rng.Shuffle(u.generatorsCount, u.shuffler)
}

func (u *Uint64) shuffler(i, j int) {
	u.generators[i], u.generators[j] = u.generators[j], u.generators[i]
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
	for i := 0; i < MaxBitsFlipped; i++ {
		var flipper = NewFlipBits(i + 1)
		for j := 0; j < MaxBitFlipperCount; j++ {
			u.Add(flipper)
		}
	}
	var random = NewRandom()
	for i := 1; i < 16; i++ {
		var byteVal = byte(i)
		u.Add(
			random,
			NewShiftedConstant(byteVal),
			NewRandomAndMask(byteVal),
			NewRandomXorMask(byteVal),
			NewRandomOrMask(byteVal),
		)
	}
	u.Shuffle()
	return u
}

func NextId() uint64 {
	return atomic.AddUint64(&id, 1)
}
