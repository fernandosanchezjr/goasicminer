package stratum

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"testing"
)

func BenchmarkWork_PlainHeader(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pw.PlainHeader()
	}
	b.StopTimer()
}

func BenchmarkWork_Clone(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	var clone *Work
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		clone = pw.Clone()
	}
	b.StopTimer()
	// method call to avoid unused variable error
	_ = clone.String()
}

func BenchmarkWork_Midstate(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	header := pw.PlainHeader()
	var ms [32]byte
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ms = utils.Midstate(header[:64])
	}
	b.StopTimer()
	// call to avoid unused variable error
	_ = ms
}

func BenchmarkWork_DoubleHash(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	header := pw.PlainHeader()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ms := utils.DoubleHash(header)
		_ = ms
	}
	b.StopTimer()
}
