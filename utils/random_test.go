package utils

import (
	log "github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestZipf(t *testing.T) {
	rng := rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 4.0, math.Pow(2, 60),
		math.MaxUint64)
	for i := 0; i < 64; i++ {
		log.WithField("rng", Nonce64(rng.Uint64())).Println("Zipf")
	}
}

func BenchmarkZipf(b *testing.B) {
	rng := rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), 4.0, math.Pow(2, 60),
		math.MaxUint64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rng.Uint64()
	}
}
