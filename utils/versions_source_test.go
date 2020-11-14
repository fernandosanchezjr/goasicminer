package utils

import (
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)

func Test_Combinations(t *testing.T) {
	vs := NewVersionSource(0x20000000, 0x1fffe000)
	var versionCount [4]Version
	for i := 0; i < 16; i++ {
		vs.Retrieve(versionCount[:])
		for _, v := range versionCount {
			log.Printf("%08x", v)
		}
	}
}

func Test_Fragment(t *testing.T) {
	wait := time.Microsecond * 17125
	count := (time.Second.Microseconds() / wait.Microseconds()) * 30
	log.WithField("count", count).Info("Rounded time")
}
