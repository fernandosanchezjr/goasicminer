package utils

import (
	"log"
	"testing"
)

func Test_Random(t *testing.T) {
	var p float64 = 8
	for i := 0; i < 64; i++ {
		log.Printf("Random: %016x", Random(p))
	}
}
