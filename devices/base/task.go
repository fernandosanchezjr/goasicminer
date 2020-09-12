package base

import (
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"sync"
)

type TaskType byte

type ITask interface {
	MarshalBinary() ([]byte, error)
	Index() int
	Update(task *stratum.Task)
	UpdateResult(tr *TaskResult, nonce uint32, versionIndex int)
	VersionsCount() int
	GetJobId() string
}

type Task struct {
	index              int
	JobId              string
	VersionRollingMask uint32
	ExtraNonce2        uint64
	NTime              uint32
	Nonce              uint32
	Versions           []uint32
	PlainHeader        [80]byte
	Pool               *stratum.Pool
	operation          uint64
	quit               chan struct{}
	wg                 sync.WaitGroup
	nextOp             chan uint64
	mtx                sync.Mutex
}

func NewTask(index int, versionsCount int) *Task {
	return &Task{index: index, Versions: make([]uint32, versionsCount), nextOp: make(chan uint64)}
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

func (t *Task) UpdateResult(tr *TaskResult, nonce uint32, versionIndex int) {
	copy(tr.PlainHeader[:], t.PlainHeader[:])
	tr.JobId = t.JobId
	tr.Version = t.Versions[versionIndex]
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
