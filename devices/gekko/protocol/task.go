package protocol

import (
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/howeyc/crc16"
)

var busyWork = []byte{0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

type Task struct {
	base.ITask
	jobId byte
	data  []byte
	busy  bool
}

func NewTask(jobId byte, versionsCount int) *Task {
	t := &Task{ITask: base.NewTask(int(jobId), versionsCount), jobId: jobId, data: make([]byte, 256)}
	t.data[0] = 0x21
	t.data[1] = 0x00
	t.data[2] = t.jobId & 0x7f
	t.data[3] = 0
	return t
}

func (t *Task) crc(dataLen byte, data []byte) {
	checkSum := crc16.ChecksumCCITTFalse(data[:dataLen-2])
	data[dataLen-2] = byte(checkSum>>8) & 0xff
	data[dataLen-1] = byte(checkSum & 0xff)
}

func (t *Task) MarshalBinary() ([]byte, error) {
	t.Lock()
	defer t.Unlock()
	return t.data[:], nil
}

func (t *Task) Update(task *stratum.Task) {
	t.Lock()
	defer t.Unlock()
	versionCount := t.VersionsCount()
	t.data[1] = byte(20 + (32 * versionCount) + 2)
	t.data[2] = t.jobId & 0x7f
	t.data[3] = byte(versionCount)
	t.data[4] = byte(task.Nbits & 0xff)
	t.data[5] = byte((task.Nbits >> 8) & 0xff)
	t.data[6] = byte((task.Nbits >> 16) & 0xff)
	t.data[7] = byte((task.Nbits >> 24) & 0xff)
	copy(t.data[8:], task.Endstate[4:])
	start := 20
	for _, midstate := range task.Midstates {
		copy(t.data[start:], midstate[:])
		start += 32
	}
	t.crc(t.data[1], t.data)
	t.ITask.Update(task)
	t.busy = false
}

func (t *Task) SetBusyWork() {
	t.Lock()
	defer t.Unlock()
	copy(t.data[4:], busyWork)
	t.data[1] = byte(4 + len(busyWork))
	t.data[2] = t.jobId & 0x7f
	t.data[3] = 1
	t.crc(t.data[1], t.data)
	t.busy = true
}

func (t *Task) IsBusyWork() bool {
	return t.busy
}
