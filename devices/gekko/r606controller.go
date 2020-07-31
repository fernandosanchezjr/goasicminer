package gekko

import (
	"bytes"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko/protocol"
	"log"
	"sync"
	"time"
)

const (
	R606BaudDiv      = 1
	R606MinFrequency = 50.0
	R606MaxFrequency = 1200
	//R606DefaultFrequency = 500
	//R606NumCores = 144
	//R606TaskLen = 54
	//R606RXLen = 7
	R606MaxJobId      = 0x7f
	R606SendQueueSize = 4
)

type R606Controller struct {
	base.IController
	base.ITaskDispatcher
	lastReset    time.Time
	frequency    float64
	chipCount    int
	tasks        []*protocol.Task
	quit         chan struct{}
	waiter       sync.WaitGroup
	sendQueue    base.TaskChan
	receiveQueue base.TaskChan
	work         base.WorkChan
}

func NewR606Controller(controller base.IController) *R606Controller {
	rc := &R606Controller{IController: controller, quit: make(chan struct{})}
	rc.ITaskDispatcher = base.NewTaskDispatcher(R606MaxJobId, rc.handleStart, rc.handleReady, rc.handleSent,
		rc.handleReceive)
	return rc
}

func (rc *R606Controller) Close() {
	waitForReader := rc.quit != nil
	if rc.quit != nil {
		close(rc.quit)
	}
	if rc.sendQueue != nil {
		close(rc.sendQueue)
	}
	if rc.work != nil {
		close(rc.work)
	}
	if waitForReader {
		rc.waiter.Wait()
	}
	if rc.ITaskDispatcher != nil {
		rc.ITaskDispatcher.Stop()
	}
	rc.IController.Close()
}

func (rc *R606Controller) Reset() error {
	log.Println("Resetting", rc.LongString())
	rc.tasks = []*protocol.Task{}
	if err := rc.PerformReset(); err != nil {
		return err
	}
	if err := rc.CountChips(); err != nil {
		return err
	} else {
		log.Println(rc.LongString(), "found", rc.chipCount, "chips")
	}
	if err := rc.SendChainInactive(); err != nil {
		return err
	}
	if err := rc.SetBaud(); err != nil {
		return err
	}
	if err := rc.InitializeTasks(); err != nil {
		return err
	}
	log.Println(rc.LongString(), "reset")
	return nil
}

func (rc *R606Controller) PerformReset() error {
	device := rc.USBDevice()
	// Reverse-engineered from cgminer with wireshark
	// FTDI Reset
	if _, err := device.Control(64, 0, 0, 0, nil); err != nil {
		return err
	}
	// FTDI Set Data
	if _, err := device.Control(64, 4, 8, 0, nil); err != nil {
		return err
	}
	// FTDI Set Baud
	if _, err := device.Control(64, 3, 0x1a, 0, nil); err != nil {
		return err
	}
	// FTDI Set Flow Control
	if _, err := device.Control(64, 2, 0, 0, nil); err != nil {
		return err
	}
	// FTDI Purge TX
	if _, err := device.Control(64, 0, 2, 0, nil); err != nil {
		return err
	}
	// FTDI Purge RX
	if _, err := device.Control(64, 0, 1, 0, nil); err != nil {
		return err
	}
	// FTDI Bitmode CB1 High
	if _, err := device.Control(64, 0xb, 0x20f2, 0, nil); err != nil {
		return err
	}
	time.Sleep(30 * time.Millisecond)
	// FTDI Bitmode CB1 Low
	if _, err := device.Control(64, 0xb, 0x20f0, 0, nil); err != nil {
		return err
	}
	time.Sleep(30 * time.Millisecond)
	// FTDI Bitmode CB1 High
	if _, err := device.Control(64, 0xb, 0x20f2, 0, nil); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (rc *R606Controller) CountChips() error {
	var buf bytes.Buffer
	cc := protocol.NewCountChips()
	data, _ := cc.MarshalBinary()
	if err := rc.Write(data); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if err := rc.Read(&buf); err != nil {
		return err
	} else {
		ccr := protocol.NewCountChipsResponse()
		if err := ccr.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
			return nil
		} else {
			rc.chipCount = len(ccr.Chips)
			return nil
		}
	}
}

func (rc *R606Controller) SendChainInactive() error {
	ci := protocol.NewChainInactive()
	cic := protocol.NewChainInactiveChip(rc.chipCount)
	if data, err := ci.MarshalBinary(); err != nil {
		return err
	} else {
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
		if err := rc.Write(data); err != nil {
			return err
		}
	}
	for i := 0; i < rc.chipCount; i++ {
		cic.SetCurrentChip(i)
		if data, err := cic.MarshalBinary(); err != nil {
			return err
		} else {
			time.Sleep(5 * time.Millisecond)
			if err := rc.Write(data); err != nil {
				return err
			}
		}
	}
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (rc *R606Controller) SetBaud() error {
	sba := protocol.NewSetBaudA(R606BaudDiv)
	sbb := protocol.NewSetBaudGateBlockMessage(R606BaudDiv)
	if data, err := sba.MarshalBinary(); err != nil {
		return err
	} else {
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Set baud
	if _, err := rc.USBDevice().Control(64, 3, 0x02, 0, nil); err != nil {
		return err
	}
	if data, err := sbb.MarshalBinary(); err != nil {
		return err
	} else {
		if err := rc.Write(data); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (rc *R606Controller) SetFrequency(frequency float64) error {
	sf := protocol.NewSetFrequency(R606MinFrequency, R606MaxFrequency, frequency)
	if sf.Frequency != rc.frequency {
		if data, err := sf.MarshalBinary(); err != nil {
			return err
		} else {
			if err := rc.Write(data); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rc *R606Controller) setupSampleWork() {
	midstates := GetWiresharkMidstates()
	rc.work = make(base.WorkChan, len(midstates))
	for _, ms := range midstates {
		rc.work <- base.NewWork(0, ms)
	}
}

func (rc *R606Controller) InitializeTasks() error {
	//rc.setupSampleWork()
	rc.sendQueue = make(base.TaskChan, R606SendQueueSize)
	rc.receiveQueue = make(base.TaskChan, R606MaxJobId)
	rc.waiter.Add(1)
	go rc.hashResponseLoop()
	rc.tasks = make([]*protocol.Task, R606MaxJobId)
	var j byte
	for j = 0; j < R606MaxJobId; j++ {
		rc.tasks[j] = protocol.NewTask(j)
	}
	go func() {
		for pos, task := range rc.tasks {
			if pos < R606SendQueueSize {
				rc.OnSent(task)
			} else {
				rc.OnReady(task)
			}
		}
	}()
	rc.ITaskDispatcher.Start()
	return nil
}

func (rc *R606Controller) handleStart(task base.ITask, dispatcher base.ITaskDispatcher) {
	defer dispatcher.OnError(task)
	if next, ok := <-rc.work; ok {
		task.Update(next.Midstate)
		rc.work <- next
	}
	dispatcher.OnReady(task)
}

func (rc *R606Controller) handleReady(task base.ITask, dispatcher base.ITaskDispatcher) {
	defer dispatcher.OnError(task)
	rc.sendQueue <- task
}

func (rc *R606Controller) handleSent(task base.ITask, dispatcher base.ITaskDispatcher) {
	defer dispatcher.OnError(task)
	if data, err := task.MarshalBinary(); err != nil {
		panic(err)
	} else {
		if err = rc.Write(data); err != nil {
			panic(err)
		}
	}
	rc.receiveQueue <- task
}

func (rc *R606Controller) handleReceive(task base.ITask, dispatcher base.ITaskDispatcher) {
	defer dispatcher.OnError(task)
	dispatcher.OnStart(task)
}

func (rc *R606Controller) hashResponseLoop() {
	var buf bytes.Buffer
	var r *protocol.TaskResponse
	var task, next base.ITask
	var ok bool
	rb := protocol.NewResponseBlock()
	readTicker := time.NewTicker(25 * time.Millisecond)
	for {
		select {
		case <-rc.quit:
			readTicker.Stop()
			rc.waiter.Done()
			return
		case <-readTicker.C:
			buf.Reset()
			if err := rc.Read(&buf); err != nil {
				log.Println("Error reading response block:", err)
				continue
			}
			if err := rb.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
				log.Println("Error decoding response block:", err)
				continue
			}
			for i := 0; i < rb.Count; i++ {
				r = rb.Responses[i]
				if next, ok = <-rc.sendQueue; ok {
					rc.OnSent(next)
				}
				task = rc.tasks[r.JobId]
				rc.OnReceived(task)
			}
		}
	}
}
