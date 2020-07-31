package protocol

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"github.com/howeyc/crc16"
)

type Task struct {
	base.ITask
	jobId byte
	data  []byte
}

func NewTask(jobId byte) *Task {
	maxLen := byte(4 + 48 + (32 * 3) + 2)
	t := &Task{ITask: base.NewTask(int(jobId)), jobId: jobId, data: make([]byte, maxLen)}
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

func (t *Task) Reverse(dataLen byte, data []byte) {
	for i, j := 4, int(dataLen-3); i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func (t *Task) Update(midstate ...utils.MidstateBytes) {
	t.data[1] = byte(4 + 48 + (32 * (len(midstate) - 1)) + 2)
	midstate[0].Reverse()
	copy(t.data[4:], midstate[0])
	t.crc(t.data[1], t.data)
}
