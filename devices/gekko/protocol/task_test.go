package protocol

import (
	"encoding/hex"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"testing"
)

func TestTask_TestEncodeNoBoost(t *testing.T) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pw.SetExtraNonce2(4)
	versions := utils.NewVersionSource(pw.Version, pw.VersionRollingMask)
	var versionMasks [4]utils.Version
	versions.Retrieve(versionMasks[:])
	pt := stratum.NewTask(1, true)
	pt.Update(pw, versionMasks[:])
	task := NewTask(0x74, 1)
	task.Update(pt)
	data, _ := task.MarshalBinary()
	hexData := hex.EncodeToString(data)
	if hexData != "2136740180ff7f1bea07101775424c5f69aa3fa0e3ff3bf3977a3140423c727f895210de0c6e467e18269b6d1936c98fd8"+
		"ede8c4197000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"+
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"+
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"+
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" {
		t.Fatal(hexData)
	}
}

func TestTask_TestEncodeBoost(t *testing.T) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pw.SetExtraNonce2(4)
	versions := utils.NewVersionSource(pw.Version, pw.VersionRollingMask)
	var versionMasks [4]utils.Version
	versions.Retrieve(versionMasks[:])
	pt := stratum.NewTask(4, true)
	pt.Update(pw, versionMasks[:])
	task := NewTask(0x74, 4)
	task.Update(pt)
	data, _ := task.MarshalBinary()
	hexData := hex.EncodeToString(data)
	if hexData != "2196740480ff7f1bea07101775424c5f69aa3fa0e3ff3bf3977a3140423c727f895210de0c6e467e18269b6d1936c98fd8"+
		"ede8c4d475dba7c1096956d79ab6b4240c70dc8ff694b1aa35a0cc8f566df9b177a6fc211bfa42e6c8b0127149c2c635e7062c70b012"+
		"0958c430b7dac7d056d5c888e1ea7a2f31209cee57cec6c0deb39fc3484818ca7dd86ddfc6598009ca244712433d5500000000000000"+
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"+
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" {
		t.Fatal(hexData)
	}
}

func BenchmarkTask_Update(b *testing.B) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	versions := utils.NewVersionSource(pw.Version, pw.VersionRollingMask)
	var versionMasks [4]utils.Version
	versions.Retrieve(versionMasks[:])
	pt := stratum.NewTask(4, true)
	pt.Update(pw, versionMasks[:])
	task := NewTask(0x75, 4)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		task.Update(pt)
	}
	b.StopTimer()
}

func TestTask_Result(t *testing.T) {
	pw, err := stratum.UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	versions := utils.NewVersionSource(pw.Version, pw.VersionRollingMask)
	var versionMasks [4]utils.Version
	versions.Retrieve(versionMasks[:])
	pt := stratum.NewTask(1, true)
	pt.Update(pw, versionMasks[:])
	pt.NTime = 0x5f2606b4
	task := NewTask(0x75, 1)
	task.Update(pt)
	tr := base.NewTaskResult()
	task.UpdateResult(tr, 0x96aa464, 0)
	tr.UpdateHeader()
	plainHeader := hex.EncodeToString(tr.PlainHeader[:])
	if plainHeader != "20000000ea2bc5140f45747839fce96b74bafe832804ed98000c817200000000000000001dccada801e1fdcd808bd3"+
		"950799e5a7174cf1ab41dea6d7a0c14da7160807085f2606b4171007ea096aa464" {
		t.Fatal(plainHeader)
	}
}
