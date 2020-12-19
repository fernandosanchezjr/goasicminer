package uint64

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestUint64_Next(t *testing.T) {
	u := NewUint64Generator()
	for i := 0; i < 1024; i++ {
		log.WithField("value", utils.Nonce64(u.Next())).Println("Next")
	}
}

func BenchmarkUint64_Next(b *testing.B) {
	u := NewUint64Generator()
	b.ResetTimer()
	var result uint64
	for i := 0; i < b.N; i++ {
		result = u.Next()
	}
	b.StopTimer()
	log.WithField("result", result).Println("Final result")
}
