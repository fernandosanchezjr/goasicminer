package utils

import (
	"log"
	"testing"
)

func Test_Combinations(t *testing.T) {
	vs := NewVersions(0x20000000, 0x1fffe000, 1, 10)
	var versionCount [4]uint32
	for i := 0; i < 16; i++ {
		vs.Retrieve(versionCount[:])
		for _, v := range versionCount {
			log.Printf("%08x", v)
		}
	}
}
