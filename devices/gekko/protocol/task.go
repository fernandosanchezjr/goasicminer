package protocol

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/howeyc/crc16"
)

type Task struct {
	base.ITask
	jobId byte
	data  []byte
}

func NewTask(jobId byte) *Task {
	maxLen := byte(20 + (32 * 4) + 2)
	t := &Task{ITask: base.NewTask(int(jobId), 4), jobId: jobId, data: make([]byte, maxLen)}
	t.data[0] = 0x21
	t.data[1] = 0x00
	t.data[2] = t.jobId & 0x7f
	t.data[3] = 0x01
	return t
}

func (t *Task) crc(dataLen byte, data []byte) {
	checkSum := crc16.ChecksumCCITTFalse(data[:dataLen-2])
	data[dataLen-2] = byte(checkSum>>8) & 0xff
	data[dataLen-1] = byte(checkSum & 0xff)
}

func (t *Task) MarshalBinary() ([]byte, error) {
	return t.data[:t.data[1]], nil
}

func (t *Task) Update(task *stratum.PoolTask) {
	t.data[3] = byte(len(task.Versions))
	t.data[1] = byte(20 + (32 * t.data[3]) + 2)
	copy(t.data[8:], task.Endstate[4:])
	start := 20
	for _, midstate := range task.Midstates {
		copy(t.data[start:], midstate[:])
		start += 32
	}
	t.crc(t.data[1], t.data)
	t.ITask.Update(task)
}
