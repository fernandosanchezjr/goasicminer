package base

import (
	"sync"
	"testing"
)

func TestTaskDispatcher(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(4)
	var started, ready, sent, received bool
	readyHandler := func(task ITask, dispatcher ITaskDispatcher) {
		dispatcher.OnSend(task)
		if !started {
			started = true
			wg.Done()
		}
	}
	sentHandler := func(task ITask, dispatcher ITaskDispatcher) {
		dispatcher.OnReceived(task, 0x1234, 0x12)
		if !ready {
			ready = true
			wg.Done()
		}
	}
	receivedHandler := func(task ITask, nonce uint32, midstate byte, dispatcher ITaskDispatcher) {
		dispatcher.OnExpired(task)
		if !sent {
			sent = true
			wg.Done()
		}
	}
	expiredHandler := func(task ITask, dispatcher ITaskDispatcher) {
		dispatcher.OnReady(task)
		if !received {
			received = true
			wg.Done()
		}
	}
	td := NewTaskDispatcher(1, readyHandler, sentHandler, receivedHandler, expiredHandler)
	td.Start()
	task := NewTask(1, 4)
	td.OnReady(task)
	wg.Wait()
	td.Stop()
}
