package node

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type Task struct {
	Work         *Work
	WorkId       uint64
	NTime        utils.NTime
	Midstates    [][32]byte
	Endstate     [16]byte
	Versions     []utils.Version
	PlainHeader  [80]byte
	Nbits        uint32
	reversed     bool
	maxMidstates int
}

func NewTask(maxMidstates int, reversed bool) *Task {
	pt := &Task{
		reversed:     reversed,
		maxMidstates: maxMidstates,
	}
	return pt
}

func (pt *Task) Update(pw *Work, versions []utils.Version) {
	pt.Work = pw
	pt.WorkId = pw.WorkId
	pt.NTime = pw.Ntime
	pt.Nbits = utils.CalculateCompactDifficulty(uint64(pw.Difficulty))
	copy(pt.PlainHeader[:], pw.PlainHeader())
	copy(pt.Endstate[:], pt.PlainHeader[64:])
	if pt.reversed {
		for j, k := 0, len(pt.Endstate)-1; j < k; j, k = j+1, k-1 {
			pt.Endstate[j], pt.Endstate[k] = pt.Endstate[k], pt.Endstate[j]
		}
	}
	pt.Versions = make([]utils.Version, pt.maxMidstates)
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
