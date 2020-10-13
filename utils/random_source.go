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
	Zipf8FF
	Zipf8Shift8
	Zipf8Shift8FF
	Zipf8Shift16
	Zipf8Shift16FF
	Zipf8Shift24
	Zipf8Shift24FF
	Zipf16
	Zipf16FF
	Zipf16Shift8
	Zipf16Shift8FF
	Zipf16Shift16
	Zipf16Shift16FF
	Zipf16Shift24
	Zipf16Shift24FF
	Zipf24
	Zipf24FF
	Zipf24Shift8
	Zipf24Shift8FF
	Zipf32
	Zipf32FF
	Zipf32Shift8
	Zipf32Shift8FF
	Zipf32Shift16
	Zipf32Shift16FF
	Zipf40
	Zipf40FF
	Zipf48
	Zipf48FF
	Zipf56
	Zipf56FF
	Xoshiro
	Xoshiro2
	Xoshiro3
	Xoshiro4
	EmbeddedZero16
	EmbeddedZero32
	EmbeddedFF16
	EmbeddedFF32
	MT
	MT2
	MT3
	MT4
	Zero
	Last
)

const MaxRNGReuse = 8

var masks = []uint64{
	0xffffffffffffffff,
	0xffffffffffffffff,
	0xffffffffffffffff,
	0xffffffffffffffff,
	0xffffffffffffffff,
	0xffffffffffffffff,
	0xffffffffffffffff,
	0xf5f5f5f5f5f5f5f5,
	0x5555555555555555,
	0xfafafafafafafafa,
	0x5a5a5a5a5a5a5a5a,
	0xa5a5a5a5a5a5a5a5,
	0x55aa55aa55aa55aa,
	0xaa55aa55aa55aa55,
}

type RandomSource struct {
	lastResult uint64
	mode       int
	bigEndian  bool
	retrieved  uint64
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
		mask: masks[0],
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
	rs.mode = rs.rng.Intn(Last)
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
	case Zipf8FF:
		rs.lastResult = rs.zipfRng8.Uint64() | 0xffffffffffffff00
	case Zipf8Shift8:
		rs.lastResult = rs.zipfRng8.Uint64() << 8
	case Zipf8Shift8FF:
		rs.lastResult = (rs.zipfRng8.Uint64() << 8) | 0xffffffffffff00ff
	case Zipf8Shift16:
		rs.lastResult = rs.zipfRng8.Uint64() << 16
	case Zipf8Shift16FF:
		rs.lastResult = (rs.zipfRng8.Uint64() << 16) | 0xffffffffff00ffff
	case Zipf8Shift24:
		rs.lastResult = rs.zipfRng8.Uint64() << 24
	case Zipf8Shift24FF:
		rs.lastResult = (rs.zipfRng8.Uint64() << 16) | 0xffffffff00ffffff
	case Zipf16:
		rs.lastResult = rs.zipfRng16.Uint64()
	case Zipf16FF:
		rs.lastResult = rs.zipfRng16.Uint64() | 0xffffffffffff0000
	case Zipf16Shift8:
		rs.lastResult = rs.zipfRng16.Uint64() << 8
	case Zipf16Shift8FF:
		rs.lastResult = rs.zipfRng16.Uint64() | 0xffffffffff0000ff
	case Zipf16Shift16:
		rs.lastResult = rs.zipfRng16.Uint64() << 16
	case Zipf16Shift16FF:
		rs.lastResult = rs.zipfRng16.Uint64() | 0xffffffff0000ffff
	case Zipf16Shift24:
		rs.lastResult = rs.zipfRng16.Uint64() << 24
	case Zipf16Shift24FF:
		rs.lastResult = rs.zipfRng16.Uint64() | 0xffffff0000ffffff
	case Zipf24:
		rs.lastResult = rs.zipfRng24.Uint64()
	case Zipf24FF:
		rs.lastResult = rs.zipfRng24.Uint64() | 0xffffffffff000000
	case Zipf24Shift8:
		rs.lastResult = rs.zipfRng24.Uint64() << 8
	case Zipf24Shift8FF:
		rs.lastResult = (rs.zipfRng24.Uint64() << 8) | 0xffffffff000000ff
	case Zipf32:
		rs.lastResult = rs.zipfRng32.Uint64()
	case Zipf32FF:
		rs.lastResult = rs.zipfRng32.Uint64() | 0xffffffff00000000
	case Zipf32Shift8:
		rs.lastResult = rs.zipfRng32.Uint64() << 8
	case Zipf32Shift8FF:
		rs.lastResult = (rs.zipfRng32.Uint64() << 8) | 0xffffff00000000ff
	case Zipf32Shift16:
		rs.lastResult = rs.zipfRng32.Uint64() << 16
	case Zipf32Shift16FF:
		rs.lastResult = (rs.zipfRng32.Uint64() << 16) | 0xffff00000000ffff
	case Zipf40:
		rs.lastResult = rs.zipfRng40.Uint64()
	case Zipf40FF:
		rs.lastResult = rs.zipfRng40.Uint64() | 0xffffff0000000000
	case Zipf48:
		rs.lastResult = rs.zipfRng48.Uint64()
	case Zipf48FF:
		rs.lastResult = rs.zipfRng48.Uint64() | 0xffff000000000000
	case Zipf56:
		rs.lastResult = rs.zipfRng56.Uint64()
	case Zipf56FF:
		rs.lastResult = rs.zipfRng56.Uint64() | 0xff00000000000000
	case Xoshiro, Xoshiro2, Xoshiro3, Xoshiro4:
		rs.lastResult = rs.xoshiroRng.Uint64()
	case EmbeddedZero16:
		rs.lastResult = rs.xoshiroRng.Uint64() & 0xffffff0000ffffff
	case EmbeddedZero32:
		rs.lastResult = rs.xoshiroRng.Uint64() & 0xffff00000000ffff
	case EmbeddedFF16:
		rs.lastResult = rs.xoshiroRng.Uint64() | 0x000000ffff000000
	case EmbeddedFF32:
		rs.lastResult = rs.xoshiroRng.Uint64() | 0x0000ffffffff0000
	case MT, MT2, MT3, MT4:
		rs.lastResult = rs.mtRng.Uint64()
	}
	if rs.retrieved >= MaxRNGReuse {
		rs.shuffle()
		rs.retrieved = 0
	}
	rs.retrieved += 1
	if rs.bigEndian {
		return SwapUint64(rs.lastResult) & rs.mask
	}
	return rs.lastResult & rs.mask
}

func (rs *RandomSource) shuffle() {
	rs.mode = rs.rng.Intn(Last)
	if rs.mode == Zero {
		rs.lastResult = 0
		rs.mode = rs.rng.Intn(2)
	}
	rs.bigEndian = rs.rng.Intn(2) == 0
	rs.mask = masks[rs.rng.Intn(len(masks))]
}

func (rs *RandomSource) Reseed() {
	rs.rng.Seed(time.Now().UnixNano())
	rs.xoshiroRng.Seed(uint64(time.Now().UnixNano()))
	rs.mtRng.Seed(uint64(time.Now().UnixNano()))
}
