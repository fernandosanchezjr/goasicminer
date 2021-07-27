package utils

import (
	"crypto/rand"
	"encoding/binary"
	"math"
)

func RandomInt64() int64 {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return int64(binary.BigEndian.Uint64(data[:]))
}

func RandomUint64() uint64 {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint64(data[:])
}

func RandomUint32() uint32 {
	var data [4]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint32(data[:])
}

func RandomIntN(max int) int {
	var data [4]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	var intN = float64(binary.BigEndian.Uint32(data[:])) / float64(math.MaxUint32)
	return int(intN * float64(max))
}
