package stratum

import (
	"encoding/hex"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"testing"
)

func TestNewTask(t *testing.T) {
	pw, err := UnmarshalTestWork()
	if pw == nil {
		t.Fatal("pool work not decoded successfully")
	}
	versions := utils.NewVersionSource(pw.Version, pw.VersionRollingMask)
	if err != nil {
		t.Fatal(err)
	}
	var versionMasks [4]utils.Version
	versions.Retrieve(versionMasks[:])
	pt := NewTask(4, true)
	pt.Update(pw, versionMasks[:])
	if len(pt.Midstates) != 4 {
		t.Fatal()
	}
	if hex.EncodeToString(pt.Endstate[:]) != "00000000ea07101775424c5f08070816" {
		t.Fatal(hex.EncodeToString(pt.Endstate[:]))
	}
	pw, err = UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pt = NewTask(2, false)
	pt.Update(pw, versionMasks[:])
	if len(pt.Midstates) != 2 {
		t.Fatal()
	}
	if hex.EncodeToString(pt.Endstate[:]) != "160807085f4c4275171007ea00000000" {
		t.Fatal(hex.EncodeToString(pt.Endstate[:]))
	}
}

func BenchmarkNewTask(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	versions := utils.NewVersionSource(pw.Version, pw.VersionRollingMask)
	var versionMasks [4]utils.Version
	versions.Retrieve(versionMasks[:])
	b.StartTimer()
	task := NewTask(4, true)
	for i := 0; i < b.N; i++ {
		task.Update(pw, versionMasks[:])
	}
	b.StopTimer()
}
