package base

type TaskType byte

const (
	Busy TaskType = iota
	Real
)

type ITask interface {
	MarshalBinary() ([]byte, error)
	Index() int
	TaskType() TaskType
}

type Task struct {
	taskType TaskType
	index    int
}

func NewTask(taskType TaskType, index int) *Task {
	return &Task{taskType: taskType, index: index}
}

func (t *Task) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (t *Task) TaskType() TaskType {
	return t.taskType
}

func (t *Task) Index() int {
	return t.index
}
