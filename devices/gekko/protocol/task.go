package protocol

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/howeyc/crc16"
)

type Task struct {
	base.ITask
	jobId    byte
	data     []byte
	busyData []byte
}

func NewTask(midstates int, jobId byte) *Task {
	dataLen := byte(54 + ((midstates - 1) * 32))
	t := &Task{ITask: base.NewTask(base.Busy, int(jobId)), jobId: jobId, data: make([]byte, dataLen),
		busyData: make([]byte, dataLen)}
	t.initialize(dataLen, t.data, base.Real)
	t.initialize(dataLen, t.busyData, base.Busy)
	return t
}

func (t *Task) initialize(dataLen byte, data []byte, taskType base.TaskType) {
	data[0] = 0x21
	data[1] = dataLen
	data[2] = t.jobId & 0x7f
	data[3] = 0x01
	if taskType == base.Busy {
		for i := 0; i < 12; i++ {
			data[8+i] = 0xff
		}
		checkSum := crc16.ChecksumCCITTFalse(data[:dataLen-2])
		data[dataLen-2] = byte(checkSum>>8) & 0xff
		data[dataLen-1] = byte(checkSum & 0xff)
	}
}

func (t *Task) MarshalBinary() ([]byte, error) {
	if t.TaskType() == base.Real {
		return t.data, nil
	} else {
		return t.busyData, nil
	}
}
