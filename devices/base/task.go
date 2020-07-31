package base

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"log"
)

type TaskType byte

type ITask interface {
	MarshalBinary() ([]byte, error)
	Index() int
	Update(midstate ...utils.MidstateBytes)
}

type Task struct {
	index int
}

func NewTask(index int) *Task {
	return &Task{index: index}
}

func (t *Task) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (t *Task) Index() int {
	return t.index
}

func (t *Task) Update(midstate ...utils.MidstateBytes) {
	log.Printf("%d unhandled midstates", len(midstate))
}
