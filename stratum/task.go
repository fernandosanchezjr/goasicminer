package stratum

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type Task struct {
	JobId              string
	VersionRollingMask uint32
	ExtraNonce2        uint64
	NTime              uint32
	Nbits              uint32
	Nonce              uint32
	Midstates          [][32]byte
	Endstate           []byte
	Versions           []uint32
	PlainHeader        [80]byte
	Pool               *Pool
	reversed           bool
}

func NewTask(pw *Work, maxMidstates int, reversed bool) *Task {
	pt := &Task{
		JobId:              pw.JobId,
		VersionRollingMask: pw.VersionRollingMask,
		ExtraNonce2:        pw.ExtraNonce2,
		NTime:              pw.Ntime,
		Nbits:              utils.CalculateCompactDifficulty(uint64(pw.Difficulty)),
		Nonce:              pw.Nonce,
		Versions:           pw.Versions(maxMidstates),
		Pool:               pw.Pool,
		reversed:           reversed,
	}
	var plainHeader = pw.PlainHeader()
	copy(pt.PlainHeader[:], plainHeader)
	var initialChunk = plainHeader[:64]
	pt.Endstate = append([]byte{}, plainHeader[64:]...)
	if reversed {
		for j, k := 0, len(pt.Endstate)-1; j < k; j, k = j+1, k-1 {
			pt.Endstate[j], pt.Endstate[k] = pt.Endstate[k], pt.Endstate[j]
		}
	}
	var versionCount = len(pt.Versions)
	var version uint32
	pt.Midstates = make([][32]byte, versionCount)
	for i := 0; i < versionCount; i++ {
		version = pt.Versions[i]
		if version == pw.Version {
			pt.Midstates[i] = utils.Midstate(initialChunk)
		} else {
			plainHeader[0] = byte((version >> 24) & 0xff)
			plainHeader[1] = byte((version >> 16) & 0xff)
			plainHeader[2] = byte((version >> 8) & 0xff)
			plainHeader[3] = byte(version & 0xff)
			pt.Midstates[i] = utils.Midstate(initialChunk)
		}
		if reversed {
			for j, k := 0, len(pt.Midstates[i])-1; j < k; j, k = j+1, k-1 {
				pt.Midstates[i][j], pt.Midstates[i][k] = pt.Midstates[i][k], pt.Midstates[i][j]
			}
		}
	}
	return pt
}

func (pt *Task) IncreaseNTime(delta uint32) {
	if delta == 0 {
		return
	}
	pt.NTime += delta
	if pt.reversed {
		endstateLen := len(pt.Endstate)
		pt.Endstate[endstateLen-4] = byte(pt.NTime & 0xff)
		pt.Endstate[endstateLen-3] = byte((pt.NTime >> 8) & 0xff)
		pt.Endstate[endstateLen-2] = byte((pt.NTime >> 16) & 0xff)
		pt.Endstate[endstateLen-1] = byte((pt.NTime >> 24) & 0xff)
	} else {
		pt.Endstate[4] = byte((pt.NTime >> 24) & 0xff)
		pt.Endstate[5] = byte((pt.NTime >> 16) & 0xff)
		pt.Endstate[6] = byte((pt.NTime >> 8) & 0xff)
		pt.Endstate[7] = byte(pt.NTime & 0xff)
	}
}
