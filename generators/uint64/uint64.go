package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync/atomic"
)

const MaxBitsFlipped = 16
const MaxBitFlipperCount = 64

var id uint64

type Generator64 interface {
	Id() uint64
	Next(previousState uint64) uint64
	Reseed()
	Clone() Generator64
}

type Uint64 struct {
	rng               *rand.Rand
	generators        []Generator64
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

func (u *Uint64) Add(generator ...Generator64) {
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

func (u *Uint64) Clone(fraction float64) *Uint64 {
	var fractionCount = int(float64(u.generatorsCount) * fraction)
	var ret = &Uint64{
		rng:               rand.New(rand.NewSource(utils.RandomInt64())),
		previousState:     0,
		previousGenerated: map[uint64]byte{},
	}
	ret.generators = make([]Generator64, fractionCount)
	ret.generatorsCount = fractionCount
	var id uint64
	var originalGenerator, clonedGenerator Generator64
	var generatorClones = map[uint64]Generator64{}
	var found bool
	for i := 0; i < fractionCount; i++ {
		if u.pos >= u.generatorsCount {
			u.pos = 0
		}
		originalGenerator = u.generators[u.pos]
		id = originalGenerator.Id()
		clonedGenerator, found = generatorClones[id]
		if !found {
			clonedGenerator = originalGenerator.Clone()
			generatorClones[id] = clonedGenerator
		}
		ret.generators[i] = clonedGenerator
		u.pos += 1
	}
	return ret
}

func NewUint64Generator() *Uint64 {
	u := NewUint64()
	var rotators = []Generator64{NewRotateLeft(), NewRotateRight()}
	for i := 0; i < 16; i++ {
		u.Add(rotators...)
	}
	var reverses = []Generator64{NewReverse(), NewReverseBytes()}
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
	shaiv := NewSHAIV()
	for i := 0; i < 16; i++ {
		var byteVal = byte(i)
		var uintVal = uint64(i)
		u.Add(
			random,
			shaiv,
			NewRandomAdd(uintVal),
			NewShiftedConstant(byteVal),
			NewUnshiftedConstant(byteVal),
			NewRandomAndMask(byteVal),
			NewUnshiftedRandomAndMask(byteVal),
			NewRandomXorMask(byteVal),
			NewUnshiftedRandomXorMask(byteVal),
			NewRandomOrMask(byteVal),
			NewUnshiftedRandomOrMask(byteVal),
		)
	}
	u.Shuffle()
	return u
}

func NextId() uint64 {
	return atomic.AddUint64(&id, 1)
}
