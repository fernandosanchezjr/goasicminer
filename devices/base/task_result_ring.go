package base

import "container/ring"

type TaskResultPool struct {
	ring *ring.Ring
}

func NewTaskResultPool(size int) *TaskResultPool {
	var ret = &TaskResultPool{ring: ring.New(size)}
	for i := 0; i < size; i++ {
		ret.ring.Value = NewTaskResult()
		ret.ring = ret.ring.Next()
	}
	return ret
}

func (trr *TaskResultPool) Next() (result *TaskResult) {
	result = (trr.ring.Value).(*TaskResult)
	trr.ring = trr.ring.Next()
	return
}
