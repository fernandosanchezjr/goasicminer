package stratum

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestNewPoolTask(t *testing.T) {
	pw, err := unmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pt := NewPoolTask(pw, binary.BigEndian)
	if len(pt.Midstates) != 4 {
		t.Fail()
	}
	if hex.EncodeToString(pt.Endstate) != "7b1dde505f2606591710b4f800000000" {
		t.Fail()
	}
	pt.IncreaseNTime(1)
	if hex.EncodeToString(pt.Endstate) != "7b1dde505f26065a1710b4f800000000" {
		t.Fail()
	}
}

func BenchmarkNewPoolTask(b *testing.B) {
	pw, err := unmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = NewPoolTask(pw, binary.BigEndian)
	}
	b.StopTimer()
}
