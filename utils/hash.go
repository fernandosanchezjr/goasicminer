package utils

import (
	"encoding/binary"
	"github.com/fernandosanchezjr/sha256-simd"
)

func DoubleHash(b []byte) [32]byte {
	first := sha256.Sum256(b)
	return sha256.Sum256(first[:])
}

func Midstate(b []byte, order binary.ByteOrder) []byte {
	return sha256.Midstate(b, order)
}
