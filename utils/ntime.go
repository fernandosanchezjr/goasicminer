package utils

import "fmt"

type NTime uint32

func (n NTime) String() string {
	return fmt.Sprintf("%08x", uint32(n))
}
