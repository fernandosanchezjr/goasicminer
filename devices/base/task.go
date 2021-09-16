package base

import (
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"sync"
)

type TaskType byte

type ITask interface {
	MarshalBinary() ([]byte, error)
	Index() int
	Update(task *node.Task)
	UpdateResult(tr *TaskResult, nonce utils.Nonce32, versionIndex int)
	VersionsCount() int
	GetWorkId() uint64
	Lock()
	Unlock()
}

type Task struct {
	Work               *node.Work
	index              int
	WorkId             uint64
	VersionRollingMask utils.Version
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

func (t *Task) Update(task *node.Task) {
	t.Work = task.Work
	t.WorkId = task.WorkId
	t.NTime = task.NTime
	copy(t.Versions, task.Versions)
	copy(t.PlainHeader[:], task.PlainHeader[:])
}

func (t *Task) UpdateResult(tr *TaskResult, nonce utils.Nonce32, versionIndex int) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	copy(tr.PlainHeader[:], t.PlainHeader[:])
	tr.Work = t.Work
	tr.WorkId = t.WorkId
	tr.Version = t.Versions[versionIndex]
	tr.Midstate = int32(versionIndex)
	tr.NTime = t.NTime
	tr.Nonce = nonce
}

func (t *Task) VersionsCount() int {
	return len(t.Versions)
}

func (t *Task) GetWorkId() uint64 {
	return t.WorkId
}

func (t *Task) Lock() {
	t.mtx.Lock()
}

func (t *Task) Unlock() {
	t.mtx.Unlock()
}
