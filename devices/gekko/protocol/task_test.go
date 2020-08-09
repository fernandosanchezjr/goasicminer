package protocol

import (
	"encoding/hex"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"testing"
)

func TestTask_UpdateBusy(t *testing.T) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pt := stratum.NewPoolTask(pw, 4, true)
	task := NewTask(0x2a)
	task.Update(pt)
	data, _ := task.MarshalBinary()
	hexData := hex.EncodeToString(data)
	if hexData != "21962a040000000000f8b410175906265f50de1d7532f9084adea58b2022ac980ebca6760fca52928bb1455a20582ffeb3"+
		"d53d069df95110f6d3d824fdab2f0688c553ff6b65824b901ada816483529f3da3e7445185232e175bb20533005b56cef052f414d874"+
		"ed338b635ede2146e597a6c0e532341b0d00729a1e5ee4dc1b1a2ffe39c11667df1dece31581f720eb5759b03555cc" {
		t.Fail()
	}
}

func BenchmarkTask_Update(b *testing.B) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	pt := stratum.NewPoolTask(pw, 4, true)
	task := NewTask(0x2a)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pt.IncreaseNTime(uint32(i))
		task.Update(pt)
	}
	b.StopTimer()
}
