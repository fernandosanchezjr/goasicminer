package utils

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func Test_Combinations(t *testing.T) {
	vs := NewVersions(0x20000000, 0x1fffe000, 1, 4)
	var versionCount [4]Version
	for i := 0; i < 16; i++ {
		vs.Retrieve(versionCount[:])
		for _, v := range versionCount {
			log.Printf("%08x", v)
		}
	}
}
