package stratum

import (
	"encoding/hex"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"testing"
)

func TestNewTask(t *testing.T) {
	pw, err := UnmarshalTestWork()
	versions := utils.NewVersions(pw.Version, pw.VersionRollingMask, 4)
	if err != nil {
		t.Fatal(err)
	}
	var versionMasks [4]uint32
	versions.Retrieve(versionMasks[:])
	pt := NewTask(pw, 4, true, versionMasks[:])
	if len(pt.Midstates) != 4 {
		t.Fatal()
	}
	if hex.EncodeToString(pt.Endstate) != "00000000ea07101775424c5f08070816" {
		t.Fatal(hex.EncodeToString(pt.Endstate))
	}
	pt.IncreaseNTime(1)
	if hex.EncodeToString(pt.Endstate) != "00000000ea07101775424c5f76424c5f" {
		t.Fatal(hex.EncodeToString(pt.Endstate))
	}
	pw, err = UnmarshalTestWork()
	if err != nil {
		t.Fatal(err)
	}
	pt = NewTask(pw, 2, false, versionMasks[:])
	if len(pt.Midstates) != 2 {
		t.Fatal()
	}
	if hex.EncodeToString(pt.Endstate) != "160807085f4c4275171007ea00000000" {
		t.Fatal(hex.EncodeToString(pt.Endstate))
	}
	pt.IncreaseNTime(1)
	if hex.EncodeToString(pt.Endstate) != "160807085f4c4276171007ea00000000" {
		t.Fatal(hex.EncodeToString(pt.Endstate))
	}
}

func BenchmarkNewTask(b *testing.B) {
	pw, err := UnmarshalTestWork()
	if err != nil {
		b.Fatal(err)
	}
	versions := utils.NewVersions(pw.Version, pw.VersionRollingMask, 4)
	var versionMasks [4]uint32
	versions.Retrieve(versionMasks[:])
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = NewTask(pw, 4, true, versionMasks[:])
	}
	b.StopTimer()
}
