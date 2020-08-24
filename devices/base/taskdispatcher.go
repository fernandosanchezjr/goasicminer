package base

import (
	"log"
	"sync"
)

type taskResponse struct {
	task     ITask
	nonce    uint32
	midstate byte
}

type TaskChan chan ITask
type TaskResponseChan chan taskResponse

type TaskFunc func(task ITask)

type ITaskDispatcher interface {
	OnReady(task ITask)
	OnSend(task ITask)
	OnReceived(task ITask, nonce uint32, midstate byte)
	OnExpired(task ITask)
	OnError(task ITask)
	Start()
	Stop()
}

type TaskHandlerFunc func(task ITask, dispatcher ITaskDispatcher)
type TaskReceivedFunc func(task ITask, nonce uint32, midstate byte, dispatcher ITaskDispatcher)

type TaskDispatcher struct {
	maxTasks       int
	ready          TaskChan
	send           TaskChan
	received       TaskResponseChan
	expired        TaskChan
	handleReady    TaskHandlerFunc
	handleSend     TaskHandlerFunc
	handleReceived TaskReceivedFunc
	handleExpired  TaskHandlerFunc
	quit           chan struct{}
	wg             sync.WaitGroup
	mtx            sync.Mutex
	started        bool
}

func NewTaskDispatcher(
	maxTasks int,
	handleReady TaskHandlerFunc,
	handleSend TaskHandlerFunc,
	handleReceived TaskReceivedFunc,
	handleExpired TaskHandlerFunc,
) *TaskDispatcher {
	return &TaskDispatcher{
		maxTasks:       maxTasks,
		ready:          make(TaskChan, maxTasks),
		send:           make(TaskChan, maxTasks),
		received:       make(TaskResponseChan, maxTasks),
		expired:        make(TaskChan, maxTasks),
		quit:           make(chan struct{}),
		handleReady:    handleReady,
		handleSend:     handleSend,
		handleReceived: handleReceived,
		handleExpired:  handleExpired,
	}
}

func (td *TaskDispatcher) OnReady(task ITask) {
	td.ready <- task
}

func (td *TaskDispatcher) OnSend(task ITask) {
	td.send <- task
}

func (td *TaskDispatcher) OnReceived(task ITask, nonce uint32, midstate byte) {
	td.received <- taskResponse{
		task:     task,
		nonce:    nonce,
		midstate: midstate,
	}
}

func (td *TaskDispatcher) OnExpired(task ITask) {
	td.expired <- task
}

func (td *TaskDispatcher) OnError(task ITask) {
	if err := recover(); err != nil {
		log.Println("TaskDispatcher error:", err)
		td.ready <- task
	}
}

func (td *TaskDispatcher) loop(taskChan TaskChan, handler TaskHandlerFunc) {
	var task ITask
	var ok bool
	for {
		select {
		case <-td.quit:
			td.wg.Done()
			return
		case task, ok = <-taskChan:
			if !ok {
				continue
			}
			handler(task, td)
		}
	}
}

func (td *TaskDispatcher) receiveLoop(taskChan TaskResponseChan, handler TaskReceivedFunc) {
	var response taskResponse
	var ok bool
	for {
		select {
		case <-td.quit:
			td.wg.Done()
			return
		case response, ok = <-taskChan:
			if !ok {
				continue
			}
			handler(response.task, response.nonce, response.midstate, td)
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
	go td.loop(td.ready, td.handleReady)
	go td.loop(td.send, td.handleSend)
	go td.receiveLoop(td.received, td.handleReceived)
	go td.loop(td.expired, td.handleExpired)
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
