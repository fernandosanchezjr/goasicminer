package base

import (
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"sync"
)

type TaskType byte

type ITask interface {
	MarshalBinary() ([]byte, error)
	Index() int
	Update(task *stratum.Task)
	UpdateResult(tr *TaskResult, nonce utils.Nonce32, versionIndex int)
	VersionsCount() int
	GetJobId() string
	Lock()
	Unlock()
}

type Task struct {
	index              int
	JobId              string
	VersionRollingMask utils.Version
	ExtraNonce2        utils.Nonce64
	NTime              utils.NTime
	Nonce              utils.Nonce32
	Versions           []utils.Version
	PlainHeader        [80]byte
	Pool               *stratum.Pool
	mtx                sync.Mutex
}

func NewTask(index int, versionsCount int) *Task {
	return &Task{index: index, Versions: make([]utils.Version, versionsCount)}
}

func (t *Task) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (t *Task) Index() int {
	return t.index
}

func (t *Task) Update(task *stratum.Task) {
	t.JobId = task.JobId
	t.VersionRollingMask = task.VersionRollingMask
	t.ExtraNonce2 = task.ExtraNonce2
	t.NTime = task.NTime
	copy(t.Versions, task.Versions)
	copy(t.PlainHeader[:], task.PlainHeader[:])
}

func (t *Task) UpdateResult(tr *TaskResult, nonce utils.Nonce32, versionIndex int) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	copy(tr.PlainHeader[:], t.PlainHeader[:])
	tr.JobId = t.JobId
	tr.Version = t.Versions[versionIndex]
	tr.Midstate = int32(versionIndex)
	tr.ExtraNonce2 = t.ExtraNonce2
	tr.NTime = t.NTime
	tr.Nonce = nonce
}

func (t *Task) VersionsCount() int {
	return len(t.Versions)
}

func (t *Task) GetJobId() string {
	return t.JobId
}

func (t *Task) Lock() {
	t.mtx.Lock()
}

func (t *Task) Unlock() {
	t.mtx.Unlock()
}
