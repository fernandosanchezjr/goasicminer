package base

import (
	"sync"
	"testing"
)

func TestTaskDispatcher(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(4)
	var started, ready, sent, received bool
	startHandler := func(task ITask, dispatcher ITaskDispatcher) {
		dispatcher.OnReady(task)
		if !started {
			started = true
			wg.Done()
		}
	}
	readyHandler := func(task ITask, dispatcher ITaskDispatcher) {
		dispatcher.OnSent(task)
		if !ready {
			ready = true
			wg.Done()
		}
	}
	sentHandler := func(task ITask, dispatcher ITaskDispatcher) {
		dispatcher.OnReceived(task)
		if !sent {
			sent = true
			wg.Done()
		}
	}
	receivedHandler := func(task ITask, dispatcher ITaskDispatcher) {
		dispatcher.OnStart(task)
		if !received {
			received = true
			wg.Done()
		}
	}
	td := NewTaskDispatcher(1, startHandler, readyHandler, sentHandler, receivedHandler)
	td.Start()
	task := NewTask(1)
	td.OnStart(task)
	wg.Wait()
	td.Stop()
}
