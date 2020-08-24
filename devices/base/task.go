package base

import (
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"sync"
	"sync/atomic"
	"time"
)

type TaskType byte

type ITask interface {
	MarshalBinary() ([]byte, error)
	Index() int
	Update(task *stratum.Task)
	Start(expiredFunc TaskFunc, timeout time.Duration)
	Stop()
	StartOperation()
	CompleteOperation()
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
	operation          uint64
	quit               chan struct{}
	wg                 sync.WaitGroup
	nextOp             chan uint64
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
}

func (t *Task) Start(expiredFunc TaskFunc, timeout time.Duration) {
	t.wg.Add(1)
	t.quit = make(chan struct{})
	go t.loop(expiredFunc, timeout)
}

func (t *Task) Stop() {
	if t.quit == nil {
		return
	}
	close(t.quit)
	t.wg.Wait()
}

func (t *Task) loop(expiredFunc TaskFunc, timeout time.Duration) {
	var op uint64
	for {
		select {
		case <-t.quit:
			t.wg.Done()
			return
		case op = <-t.nextOp:
			time.Sleep(timeout)
			if atomic.LoadUint64(&t.operation) == op {
				expiredFunc(t)
			}
		}
	}
}

func (t *Task) StartOperation() {
	t.nextOp <- atomic.AddUint64(&t.operation, 1)
}

func (t *Task) CompleteOperation() {
	atomic.AddUint64(&t.operation, 1)
}
