package utils

import (
	"crypto/rand"
	"encoding/binary"
	"math"
)

func Random(p float64) uint64 {
	var randBuf [4]byte
	if _, err := rand.Read(randBuf[:]); err != nil {
		panic(err)
	}
	rnd := binary.LittleEndian.Uint32(randBuf[:])
	bias := math.Pow(float64(rnd)/float64(math.MaxUint32), p)
	return uint64(math.Round(float64(math.MaxUint64) * bias))
}
