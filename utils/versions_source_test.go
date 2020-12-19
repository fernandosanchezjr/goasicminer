package utils

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
	"testing"
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

func BenchmarkVersionSource_Shuffle(b *testing.B) {
	vs := NewVersionSource(0x20000000, 0x1fffe000)
	rng := rand.New(rand.NewSource(RandomInt64()))
	for i := 0; i < b.N; i++ {
		vs.Shuffle(rng)
	}
}
