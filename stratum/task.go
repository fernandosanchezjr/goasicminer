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
	Midstates          [][32]byte
	Endstate           [16]byte
	Versions           []uint32
	PlainHeader        [80]byte
	reversed           bool
	maxMidstates       int
}

func NewTask(maxMidstates int, reversed bool) *Task {
	pt := &Task{
		reversed:     reversed,
		maxMidstates: maxMidstates,
	}
	return pt
}

func (pt *Task) Update(pw *Work, versions []uint32) {
	pt.JobId = pw.JobId
	pt.VersionRollingMask = pw.VersionRollingMask
	pt.ExtraNonce2 = pw.ExtraNonce2
	pt.NTime = pw.Ntime
	pt.Nbits = utils.CalculateCompactDifficulty(uint64(pw.Difficulty))
	copy(pt.PlainHeader[:], pw.PlainHeader())
	copy(pt.Endstate[:], pt.PlainHeader[64:])
	if pt.reversed {
		for j, k := 0, len(pt.Endstate)-1; j < k; j, k = j+1, k-1 {
			pt.Endstate[j], pt.Endstate[k] = pt.Endstate[k], pt.Endstate[j]
		}
	}
	if !pw.VersionRolling {
		pt.Versions = []uint32{pw.Version}
		pt.Midstates = [][32]byte{utils.Midstate(pt.PlainHeader[:64])}
	} else {
		pt.Versions = make([]uint32, pt.maxMidstates)
		pt.Midstates = make([][32]byte, pt.maxMidstates)
		copy(pt.Versions, versions)
		for i, version := range pt.Versions {
			if version == 0x0 {
				pt.Versions = pt.Versions[0:i]
				pt.Midstates = pt.Midstates[0:i]
				break
			} else if version == pw.Version {
				pt.Midstates[i] = utils.Midstate(pt.PlainHeader[:64])
			} else {
				pt.PlainHeader[0] = byte((version >> 24) & 0xff)
				pt.PlainHeader[1] = byte((version >> 16) & 0xff)
				pt.PlainHeader[2] = byte((version >> 8) & 0xff)
				pt.PlainHeader[3] = byte(version & 0xff)
				pt.Midstates[i] = utils.Midstate(pt.PlainHeader[:64])
			}
			if pt.reversed {
				for j, k := 0, len(pt.Midstates[i])-1; j < k; j, k = j+1, k-1 {
					pt.Midstates[i][j], pt.Midstates[i][k] = pt.Midstates[i][k], pt.Midstates[i][j]
				}
			}
		}
	}
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
