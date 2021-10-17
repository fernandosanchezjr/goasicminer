package protocol

import "container/list"

type TaskPool struct {
	Size    int
	Tasks   []*Task
	List    *list.List
	Current *list.Element
}

func NewTaskPool(size int, midstates int) *TaskPool {
	var ret = &TaskPool{
		Size:  size,
		Tasks: make([]*Task, size),
		List:  list.New(),
	}
	for i := 0; i < size; i++ {
		var task = NewTask(byte(i), midstates)
		ret.Tasks[i] = task
		ret.List.PushBack(task)
	}
	ret.Current = ret.List.Front()
	return ret
}

func (tl *TaskPool) GetTask(index int) *Task {
	if index >= tl.Size || index < 0 {
		return nil
	}
	return tl.Tasks[index]
}

func (tl *TaskPool) Next(steps int) (next *Task, last bool) {
	if tl.Current != nil {
		next = (tl.Current.Value).(*Task)
	}
	for i := 0; i < steps && tl.Current != nil; i++ {
		tl.Current = tl.Current.Next()
	}
	last = tl.Current == nil
	if last {
		tl.Current = tl.List.Front()
	}
	return
}
