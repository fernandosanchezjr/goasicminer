package utils

import "fmt"

type Nonce64 uint64
type Nonce32 uint32

func (n Nonce64) String() string {
	return fmt.Sprintf("%016x", uint64(n))
}

func (n Nonce32) String() string {
	return fmt.Sprintf("%08x", uint32(n))
}
