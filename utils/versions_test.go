package utils

import (
	"log"
	"testing"
)

func Test_Combinations(t *testing.T) {
	log.Printf("%08x", (0x1fffe000 & ^0x20000000))
	log.Printf("%08x", (0x20000000 & ^0x20000000))
	vs := NewVersions(0x20000000, 0x1fffe000, 4)
	var versionCount [4]uint32
	for i := 0; i < 16; i++ {
		vs.Retrieve(versionCount[:])
		for _, v := range versionCount {
			log.Printf("%08x", v)
		}
	}
}
