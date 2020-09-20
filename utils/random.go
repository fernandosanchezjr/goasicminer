package utils

import (
	rand2 "bitbucket.org/MaVo159/rand"
	"crypto/rand"
	"encoding/binary"
	"math"
	"time"
)

func init() {
	SeedMT()
}

func Random(p float64) uint64 {
	var randBuf [4]byte
	if _, err := rand.Read(randBuf[:]); err != nil {
		panic(err)
	}
	rnd := binary.LittleEndian.Uint32(randBuf[:])
	bias := math.Pow(float64(rnd)/float64(math.MaxUint32), p)
	return uint64(math.Round(float64(math.MaxUint64) * bias))
}

func RandomUint64() uint64 {
	return rand2.Uint64()
}

func SeedMT() {
	rand2.Seed(uint64(time.Now().UnixNano()))
}

func RandRange(start, end int) int {
	return start + rand2.Intn(end-start)
}
