package protocol

import (
	"log"
	"testing"
)

func TestTaskPool_GetTask(t *testing.T) {
	var tp = NewTaskPool(4, 4)
	var next, last = tp.Next(4)
	if next == nil {
		t.Fail()
	}
	if !last {
		t.Fail()
	}
}

func TestTaskPool_GetTaskSingle(t *testing.T) {
	var tp = NewTaskPool(4, 4)
	for i := 0; i < 4; i++ {
		var next, last = tp.Next(1)
		if i < 3 && last {
			t.Fail()
		}
		if next == nil {
			t.Fail()
		}
		log.Println(next.GetWorkId())
	}
}
