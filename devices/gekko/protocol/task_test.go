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
	versions := utils.NewVersions(pw.Version, pw.VersionRollingMask, 4, 4)
	var versionMasks [4]uint32
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
	versions := utils.NewVersions(pw.Version, pw.VersionRollingMask, 4, 4)
	var versionMasks [4]uint32
	versions.Retrieve(versionMasks[:])
	pt := stratum.NewTask(4, true)
	pt.Update(pw, versionMasks[:])
	task := NewTask(0x74, 4)
	task.Update(pt)
	data, _ := task.MarshalBinary()
	hexData := hex.EncodeToString(data)
	if hexData != "2196740480ff7f1bea07101775424c5f69aa3fa0e3ff3bf3977a3140423c727f895210de0c6e467e18269b6d1936c98fd8"+
		"ede8c4e1ce4216a7de12fb74e694416ba7b3ed12ea12b56bbbac1fa7d97d1ee2367ee3b01463e804106098afdb3734b57f7d72b47001"+
		"b5fce77421f8c1b0ad71201c0ce1caa78738d0fbbc3f2e01e09bd2c1a5de05c902a909acdd94c6b29fbded9d007caf00000000000000"+
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
	versions := utils.NewVersions(pw.Version, pw.VersionRollingMask, 4, 4)
	var versionMasks [4]uint32
	versions.Retrieve(versionMasks[:])
	pt := stratum.NewTask(4, true)
	pt.Update(pw, versionMasks[:])
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
	versions := utils.NewVersions(pw.Version, pw.VersionRollingMask, 4, 4)
	var versionMasks [4]uint32
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
