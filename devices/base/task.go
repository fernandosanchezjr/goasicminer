package base

import (
	"github.com/fernandosanchezjr/goasicminer/stratum"
)

type TaskType byte

type ITask interface {
	MarshalBinary() ([]byte, error)
	Index() int
	Update(task *stratum.PoolTask)
}

type Task struct {
	index              int
	JobId              string
	VersionRollingMask uint32
	ExtraNonce2        uint64
	NTime              uint32
	Nonce              uint32
	Versions           []uint32
	Pool               *stratum.Pool
}

func NewTask(index int, versionsCount int) *Task {
	return &Task{index: index, Versions: make([]uint32, versionsCount)}
}

func (t *Task) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (t *Task) Index() int {
	return t.index
}

func (t *Task) Update(task *stratum.PoolTask) {
	t.JobId = task.JobId
	t.VersionRollingMask = task.VersionRollingMask
	t.ExtraNonce2 = task.ExtraNonce2
	t.NTime = task.NTime
	copy(t.Versions, task.Versions)
}
