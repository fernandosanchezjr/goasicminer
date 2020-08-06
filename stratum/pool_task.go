package stratum

import (
	"encoding/binary"
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type PoolTask struct {
	JobId              string
	VersionRollingMask uint32
	ExtraNonce2        uint64
	NTime              uint32
	Nonce              uint32
	Midstates          [][]byte
	Endstate           []byte
	Versions           []uint32
	Pool               *Pool
}

func NewPoolTask(pw *PoolWork, order binary.ByteOrder) *PoolTask {
	pt := &PoolTask{
		JobId:              pw.JobId,
		VersionRollingMask: pw.VersionRollingMask,
		ExtraNonce2:        pw.ExtraNonce2,
		NTime:              pw.Ntime,
		Nonce:              pw.Nonce,
		Versions:           pw.Versions(),
		Pool:               pw.Pool,
	}
	plainHeader := pw.PlainHeader()
	pt.Endstate = plainHeader[64:]
	pt.Midstates = make([][]byte, len(pt.Versions))
	for pos, v := range pt.Versions {
		if v == pw.Version {
			pt.Midstates[pos] = utils.Midstate(plainHeader[:64], order)
		} else {
			plainHeader[0] = byte((v >> 24) & 0xff)
			plainHeader[1] = byte((v >> 16) & 0xff)
			plainHeader[2] = byte((v >> 8) & 0xff)
			plainHeader[3] = byte(v & 0xff)
			pt.Midstates[pos] = utils.Midstate(plainHeader[:64], order)
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
