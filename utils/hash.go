package utils

import (
	"github.com/minio/sha256-simd"
)

func DoubleHash(b []byte) [32]byte {
	first := sha256.Sum256(b)
	return sha256.Sum256(first[:])
}
