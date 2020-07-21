package base

import (
	"sync"
)

type TaskChan chan ITask
type ITaskDispatcher interface {
	OnStart(task ITask)
	OnReady(task ITask)
	OnSent(task ITask)
	OnReceived(task ITask)
	OnError(task ITask)
	Start()
	Stop()
}

type TaskFunc func(task ITask, dispatcher ITaskDispatcher)

type TaskDispatcher struct {
	maxTasks       int
	start          TaskChan
	ready          TaskChan
	sent           TaskChan
	received       TaskChan
	handleStart    TaskFunc
	handleReady    TaskFunc
	handleSent     TaskFunc
	handleReceived TaskFunc
	quit           chan struct{}
	wg             sync.WaitGroup
	mtx            sync.Mutex
	started        bool
}

func NewTaskDispatcher(maxTasks int, handleStart, handleReady, handleSent, handleReceived TaskFunc) *TaskDispatcher {
	return &TaskDispatcher{
		maxTasks:       maxTasks,
		start:          make(TaskChan, maxTasks),
		ready:          make(TaskChan, maxTasks),
		sent:           make(TaskChan, maxTasks),
		received:       make(TaskChan, maxTasks),
		quit:           make(chan struct{}),
		handleStart:    handleStart,
		handleReady:    handleReady,
		handleSent:     handleSent,
		handleReceived: handleReceived,
	}
}

func (td *TaskDispatcher) OnStart(task ITask) {
	td.start <- task
}

func (td *TaskDispatcher) OnReady(task ITask) {
	td.ready <- task
}

func (td *TaskDispatcher) OnSent(task ITask) {
	td.sent <- task
}

func (td *TaskDispatcher) OnReceived(task ITask) {
	td.received <- task
}

func (td *TaskDispatcher) OnError(task ITask) {
	if err := recover(); err != nil {
		td.start <- task
	}
}

func (td *TaskDispatcher) loop(taskChan TaskChan, handler TaskFunc) {
	var task ITask
	for {
		select {
		case <-td.quit:
			td.wg.Done()
			return
		case task = <-taskChan:
			handler(task, td)
		}
	}
}

func (td *TaskDispatcher) Start() {
	td.mtx.Lock()
	defer td.mtx.Unlock()
	if td.started {
		return
	}
	td.started = true
	td.wg.Add(4)
	go td.loop(td.start, td.handleStart)
	go td.loop(td.ready, td.handleReady)
	go td.loop(td.sent, td.handleSent)
	go td.loop(td.received, td.handleReceived)
}

func (td *TaskDispatcher) Stop() {
	td.mtx.Lock()
	defer td.mtx.Unlock()
	if !td.started {
		return
	}
	close(td.quit)
	td.started = false
	td.wg.Wait()
	td.quit = make(chan struct{})
}
