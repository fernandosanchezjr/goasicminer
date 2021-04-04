package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync/atomic"
)

const MaxBitsFlipped = 8
const MaxBitFlipperCount = 24

var id uint64

type Uint64 struct {
	rng               *rand.Rand
	generators        []utils.Generator64
	generatorsCount   int
	previousState     uint64
	previousGenerated map[uint64]byte
	ri                *utils.RandomIndex
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
	var nextPos = u.ri.Next(u.rng)
	u.previousState = u.generators[nextPos].Next(u.previousState)
	return u.previousState
}

func (u *Uint64) Shuffle() {
	u.rng.Shuffle(u.generatorsCount, u.shuffler)
}

func (u *Uint64) shuffler(i, j int) {
	u.generators[i], u.generators[j] = u.generators[j], u.generators[i]
}

func (u *Uint64) Reset() {
	u.ri.Reset()
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
	var rl, rr = NewRotateLeft(), NewRotateRight()
	var random = NewRandom()
	for i := 1; i < 16; i++ {
		var byteVal = byte(i)
		u.Add(
			random,
			NewShiftedConstant(byteVal),
			NewRandomAndMask(byteVal),
			NewRandomXorMask(byteVal),
			NewRandomOrMask(byteVal),
			rl, rr,
		)
	}
	u.Shuffle()
	u.ri = utils.NewRandomIndex(u.generatorsCount)
	return u
}

func NextId() uint64 {
	return atomic.AddUint64(&id, 1)
}
