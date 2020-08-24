package base

import (
	"testing"
	"time"
)

func TestTaskCompletion(t *testing.T) {
	task := NewTask(1, 1)
	var expired bool
	expiredfunc := func(_ ITask) {
		expired = true
	}
	task.Start(expiredfunc, time.Duration(100*time.Millisecond))
	task.StartOperation()
	time.Sleep(10 * time.Millisecond)
	task.CompleteOperation()
	time.Sleep(150 * time.Millisecond)
	if expired == true {
		t.Fail()
	}
}

func TestTaskExpiration(t *testing.T) {
	task := NewTask(1, 1)
	var expired bool
	expiredfunc := func(_ ITask) {
		expired = true
	}
	task.Start(expiredfunc, time.Duration(100*time.Millisecond))
	task.StartOperation()
	time.Sleep(150 * time.Millisecond)
	if expired != true {
		t.Fail()
	}
}
