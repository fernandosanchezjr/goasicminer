package utils

import "fmt"

type Version uint32

func (v Version) String() string {
	return fmt.Sprintf("%08x", uint32(v))
}
