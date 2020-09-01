package protocol

import (
	"encoding/hex"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"testing"
)

func TestTask_UpdateBusy(t *testing.T) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pw.SetExtraNonce2(5)
	pt := stratum.NewTask(pw, 4, true)
	task := NewTask(0x75, 1)
	task.Update(pt)
	data, _ := task.MarshalBinary()
	hexData := hex.EncodeToString(data)
	if hexData != "2136750100000000ea07101775424c5f8b4382dcf25f41a9e5fcee0f708cb3f1de287b3e1e78a58eb128539f8e1b866a2393d62ed8ae" {
		t.Fatal(hexData)
	}
}

func TestTask_TestEncoded(t *testing.T) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pw.SetExtraNonce2(4)
	pt := stratum.NewTask(pw, 1, true)
	task := NewTask(0x74, 1)
	task.Update(pt)
	data, _ := task.MarshalBinary()
	hexData := hex.EncodeToString(data)
	if hexData != "2136740100000000ea07101775424c5f69aa3fa0e3ff3bf3977a3140423c727f895210de0c6e467e18269b6d1936c98fd8ede8c4213c" {
		t.Fatal(hexData)
	}
}

func BenchmarkTask_Update(b *testing.B) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	pt := stratum.NewTask(pw, 4, true)
	task := NewTask(0x75, 4)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pt.IncreaseNTime(uint32(i))
		task.Update(pt)
	}
	b.StopTimer()
}

func TestTask_Result(t *testing.T) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pt := stratum.NewTask(pw, 1, true)
	pt.NTime = 0x5f2606b4
	task := NewTask(0x75, 1)
	task.Update(pt)
	tr := base.NewTaskResult()
	task.UpdateResult(tr, 0x96aa464, 0)
	tr.UpdateHeader()
	plainHeader := hex.EncodeToString(tr.PlainHeader[:])
	if plainHeader != "20000000ea2bc5140f45747839fce96b74bafe832804ed98000c817200000000000000001dccada801e1fdcd808bd3950799e5a7174cf1ab41dea6d7a0c14da7160807085f2606b4171007ea096aa464" {
		t.Fatal(plainHeader)
	}
}
