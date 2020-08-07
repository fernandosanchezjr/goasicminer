package stratum

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type PoolTask struct {
	JobId              string
	VersionRollingMask uint32
	ExtraNonce2        uint64
	NTime              uint32
	Nonce              uint32
	Midstates          [][32]byte
	Endstate           []byte
	Versions           []uint32
	Pool               *Pool
}

func NewPoolTask(pw *PoolWork, max int) *PoolTask {
	pt := &PoolTask{
		JobId:              pw.JobId,
		VersionRollingMask: pw.VersionRollingMask,
		ExtraNonce2:        pw.ExtraNonce2,
		NTime:              pw.Ntime,
		Nonce:              pw.Nonce,
		Versions:           pw.Versions(max),
		Pool:               pw.Pool,
	}
	var plainHeader = pw.PlainHeader()
	var initialChunk = plainHeader[:64]
	pt.Endstate = plainHeader[64:]
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
	}
	return pt
}

func (pt *PoolTask) IncreaseNTime(delta uint32) {
	pt.NTime += delta
	pt.Endstate[4] = byte((pt.NTime >> 24) & 0xff)
	pt.Endstate[5] = byte((pt.NTime >> 16) & 0xff)
	pt.Endstate[6] = byte((pt.NTime >> 8) & 0xff)
	pt.Endstate[7] = byte(pt.NTime & 0xff)
}
