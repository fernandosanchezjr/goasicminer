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
	Zipf8
	Zipf16
	Zipf24
	Zipf32
	Zipf40
	Zipf48
	Zipf56
	Xoshiro
	MT
	Last
)

var masks = []uint64{
	0xffffffffffffffff,
	0x5555555555555555,
	0xaaaaaaaaaaaaaaaa,
	0xa5a5a5a5a5a5a5a5,
	0x5a5a5a5a5a5a5a5a,
	0x0000ffffffff0000,
	0x000000ffff000000,
}

type RandomSource struct {
	lastResult uint64
	mode       int
	bigEndian  bool
	mask       uint64
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
		mask: 0xffffffffffffffff,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
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
	case Random:
		rs.lastResult = rs.rng.Uint64() & rs.mask
	case Zipf8:
		rs.lastResult = rs.zipfRng8.Uint64()
	case Zipf16:
		rs.lastResult = rs.zipfRng16.Uint64()
	case Zipf24:
		rs.lastResult = rs.zipfRng24.Uint64() & rs.mask
	case Zipf32:
		rs.lastResult = rs.zipfRng32.Uint64() & rs.mask
	case Zipf40:
		rs.lastResult = rs.zipfRng40.Uint64() & rs.mask
	case Zipf48:
		rs.lastResult = rs.zipfRng48.Uint64() & rs.mask
	case Zipf56:
		rs.lastResult = rs.zipfRng56.Uint64() & rs.mask
	case Xoshiro:
		rs.lastResult = rs.xoshiroRng.Uint64() & rs.mask
	case MT:
		rs.lastResult = rs.mtRng.Uint64() & rs.mask
	}
	if rs.bigEndian {
		return SwapUint64(rs.lastResult)
	}
	return rs.lastResult
}

func (rs *RandomSource) Shuffle() {
	rs.mode = rs.rng.Intn(Last)
	rs.bigEndian = rs.rng.Intn(1000)%2 == 0
	rs.mask = masks[rs.rng.Intn(len(masks))]
}
