package utils

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func Test_Random(t *testing.T) {
	var p float64 = 8
	for i := 0; i < 64; i++ {
		log.Printf("Random: %016x", Random(p))
	}
}

func Test_RandomUint64(t *testing.T) {
	for i := 0; i < 64; i++ {
		log.Printf("Random: %016x", RandomUint64())
	}
}
