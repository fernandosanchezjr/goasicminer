package utils

import (
	"gonum.org/v1/gonum/mathext/prng"
	"math"
	"math/rand"
	"time"
)

const (
	Add int = iota
	Sub
	Random
	Random2
	Random3
	Random4
	Zipf8
	Zipf8Shift8
	Zipf8Shift16
	Zipf8Shift24
	Zipf16
	Zipf16Shift8
	Zipf16Shift16
	Zipf16Shift24
	Zipf24
	Zipf32
	Zipf32Shift8
	Zipf32Shift16
	Zipf40
	Zipf48
	Zipf56
	Xoshiro
	Xoshiro2
	Xoshiro3
	Xoshiro4
	EmbeddedZero16
	EmbeddedZero32
	MT
	MT2
	MT3
	MT4
	Zero
	Last
)

type RandomSource struct {
	lastResult uint64
	mode       int
	bigEndian  bool
	rng        *rand.Rand
	zipfRng8   *rand.Zipf
	zipfRng16  *rand.Zipf
	zipfRng24  *rand.Zipf
	zipfRng32  *rand.Zipf
	zipfRng40  *rand.Zipf
	zipfRng48  *rand.Zipf
	zipfRng56  *rand.Zipf
	xoshiroRng *prng.Xoshiro256starstar
	mtRng      *prng.MT19937_64
}

func NewRandomSource() *RandomSource {
	rs := &RandomSource{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		zipfRng8: rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 2.0, math.Pow(2, 8),
			math.MaxUint64),
		zipfRng16: rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 2.0, math.Pow(2, 16),
			math.MaxUint64),
		zipfRng24: rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 2.0, math.Pow(2, 24),
			math.MaxUint64),
		zipfRng32: rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 2.0, math.Pow(2, 32),
			math.MaxUint64),
		zipfRng40: rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 2.0, math.Pow(2, 40),
			math.MaxUint64),
		zipfRng48: rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 2.0, math.Pow(2, 48),
			math.MaxUint64),
		zipfRng56: rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 2.0, math.Pow(2, 56),
			math.MaxUint64),
		xoshiroRng: prng.NewXoshiro256starstar(uint64(time.Now().UnixNano())),
		mtRng:      prng.NewMT19937_64(),
	}
	rs.mtRng.Seed(uint64(time.Now().UnixNano()))
	return rs
}

func (rs *RandomSource) Uint64() uint64 {
	switch rs.mode {
	case Add:
		rs.lastResult += 1
	case Sub:
		rs.lastResult -= 1
	case Random, Random2, Random3, Random4:
		rs.lastResult = rs.rng.Uint64()
	case Zipf8:
		rs.lastResult = rs.zipfRng8.Uint64()
	case Zipf8Shift8:
		rs.lastResult = rs.zipfRng8.Uint64() << 8
	case Zipf8Shift16:
		rs.lastResult = rs.zipfRng8.Uint64() << 16
	case Zipf8Shift24:
		rs.lastResult = rs.zipfRng8.Uint64() << 24
	case Zipf16:
		rs.lastResult = rs.zipfRng16.Uint64()
	case Zipf16Shift8:
		rs.lastResult = rs.zipfRng16.Uint64() << 8
	case Zipf16Shift16:
		rs.lastResult = rs.zipfRng16.Uint64() << 16
	case Zipf16Shift24:
		rs.lastResult = rs.zipfRng16.Uint64() << 24
	case Zipf24:
		rs.lastResult = rs.zipfRng24.Uint64()
	case Zipf32:
		rs.lastResult = rs.zipfRng32.Uint64()
	case Zipf32Shift8:
		rs.lastResult = rs.zipfRng32.Uint64() << 8
	case Zipf32Shift16:
		rs.lastResult = rs.zipfRng32.Uint64() << 16
	case Zipf40:
		rs.lastResult = rs.zipfRng40.Uint64()
	case Zipf48:
		rs.lastResult = rs.zipfRng48.Uint64()
	case Zipf56:
		rs.lastResult = rs.zipfRng56.Uint64()
	case Xoshiro, Xoshiro2, Xoshiro3, Xoshiro4:
		rs.lastResult = rs.xoshiroRng.Uint64()
	case EmbeddedZero16:
		rs.lastResult = rs.xoshiroRng.Uint64() & 0xffffff0000ffffff
	case EmbeddedZero32:
		rs.lastResult = rs.xoshiroRng.Uint64() & 0xffff00000000ffff
	case MT, MT2, MT3, MT4:
		rs.lastResult = rs.mtRng.Uint64()
	}
	if rs.bigEndian {
		return SwapUint64(rs.lastResult)
	}
	return rs.lastResult
}

func (rs *RandomSource) Shuffle() {
	rs.mode = rs.rng.Intn(Last)
	if rs.mode == Zero {
		rs.lastResult = 0
		rs.mode = rs.rng.Intn(2)
	}
	rs.bigEndian = rs.rng.Intn(1000)%2 == 0
}

func (rs *RandomSource) Reseed() {
	rs.rng.Seed(time.Now().UnixNano())
	rs.xoshiroRng.Seed(uint64(time.Now().UnixNano()))
	rs.mtRng.Seed(uint64(time.Now().UnixNano()))
}
