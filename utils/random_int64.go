package utils

import (
	"crypto/rand"
	"encoding/binary"
)

func RandomInt64() int64 {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return int64(binary.BigEndian.Uint64(data[:]))
}
