package gekko

import (
	"bytes"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko/protocol"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	R606BaudDiv          = 1
	R606MinFrequency     = 200
	R606MaxFrequency     = 1200
	R606DefaultFrequency = 200
	R606NumCores         = 114
	//R606RXLen = 7
	R606MaxJobId = 0x7f
	//R606SendQueueSize    = 4
	R606NtimeRollSeconds = 60
	R606MidstateCount    = 4
	R606WaitFactor       = 0.5
	R606ReadTimeout      = time.Duration(5 * time.Millisecond)
)

type R606Controller struct {
	base.IController
	base.ITaskDispatcher
	lastReset        time.Time
	frequency        float64
	chipCount        int
	tasks            []*protocol.Task
	quitRead         chan struct{}
	quitWrite        chan struct{}
	waiter           sync.WaitGroup
	sendQueue        base.TaskChan
	currentPoolWork  *stratum.Work
	currentPoolTask  *stratum.Task
	nTimeOffset      uint32
	taskMtx          sync.Mutex
	fullscanDuration time.Duration
	maxTaskWait      time.Duration
}

func NewR606Controller(controller base.IController) *R606Controller {
	rc := &R606Controller{IController: controller, quitRead: make(chan struct{}), quitWrite: make(chan struct{}),
		frequency: R606DefaultFrequency}
	rc.ITaskDispatcher = base.NewTaskDispatcher(R606MaxJobId, rc.handleReady, rc.handleSend, rc.handleReceived,
		rc.handleExpired)
	return rc
}

func (rc *R606Controller) Close() {
	waitForReader := rc.quitRead != nil
	for _, t := range rc.tasks {
		t.Stop()
	}
	if rc.quitRead != nil {
		close(rc.quitRead)
	}
	if rc.quitWrite != nil {
		close(rc.quitWrite)
	}
	if rc.sendQueue != nil {
		close(rc.sendQueue)
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
	for i := 0; i < rc.chipCount; i++ {
		if err := rc.SetFrequency(R606DefaultFrequency, byte(i)); err != nil {
			return err
		}
	}
	rc.SetTiming()
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

func (rc *R606Controller) SetFrequency(frequency float64, asicId byte) error {
	sf := protocol.NewSetFrequency(R606MinFrequency, R606MaxFrequency, frequency, asicId)
	if sf.Frequency != rc.frequency {
		if data, err := sf.MarshalBinary(); err != nil {
			return err
		} else {
			if err := rc.Write(data); err != nil {
				return err
			}
			rc.SetTiming()
		}
	}
	return nil
}

func (rc *R606Controller) SetTiming() {
	hashRate := float64(rc.chipCount) * rc.frequency * R606NumCores * 1000000.0
	fullScanMicroSeconds := 1000000.0 * (float64(0xffffffff) / hashRate)
	rc.fullscanDuration = time.Duration(time.Duration(fullScanMicroSeconds*1000.0) * time.Nanosecond)
	rc.maxTaskWait = time.Duration((R606WaitFactor * 4) * float64(rc.fullscanDuration))
	minTaskWait := time.Duration(1 * time.Microsecond)
	maxTaskWait := 3 * rc.fullscanDuration
	if rc.maxTaskWait < minTaskWait {
		rc.maxTaskWait = minTaskWait
	}
	if rc.maxTaskWait > maxTaskWait {
		rc.maxTaskWait = maxTaskWait
	}
	log.Println("Full scan time:", rc.fullscanDuration)
	log.Println("Max task wait:", rc.maxTaskWait)
}

func (rc *R606Controller) InitializeTasks() error {
	var task *protocol.Task
	rc.sendQueue = make(base.TaskChan, R606MaxJobId)
	rc.waiter.Add(2)
	go rc.writeLoop()
	go rc.readLoop()
	rc.tasks = make([]*protocol.Task, R606MaxJobId)
	var j byte
	for j = 0; j < R606MaxJobId; j++ {
		task = protocol.NewTask(j)
		task.SetBusyWork()
		rc.tasks[j] = task
	}
	return nil
}

func (rc *R606Controller) handleReady(task base.ITask, dispatcher base.ITaskDispatcher) {
	rc.taskMtx.Lock()
	defer rc.taskMtx.Unlock()
	defer dispatcher.OnError(task)
	if rc.nTimeOffset == R606NtimeRollSeconds {
		rc.currentPoolWork.SetExtraNonce2(rc.currentPoolWork.ExtraNonce2 + 1)
		rc.currentPoolTask = stratum.NewTask(rc.currentPoolWork, R606MidstateCount, true)
		rc.nTimeOffset = 0
	}
	rc.currentPoolTask.IncreaseNTime(rc.nTimeOffset)
	atomic.AddUint32(&rc.nTimeOffset, 1)
	task.Update(rc.currentPoolTask)
	rc.sendQueue <- task
}

func (rc *R606Controller) handleSend(task base.ITask, dispatcher base.ITaskDispatcher) {
	defer dispatcher.OnError(task)
	if data, err := task.MarshalBinary(); err != nil {
		panic(err)
	} else {
		if err = rc.Write(data); err != nil {
			panic(err)
		}
	}
	task.StartOperation()
}

func (rc *R606Controller) handleReceived(task base.ITask, nonce uint32, midstate byte, dispatcher base.ITaskDispatcher) {
	defer dispatcher.OnError(task)
	log.Println("Received nonce", nonce, "for midstate", midstate, "from task", task.Index())
}

func (rc *R606Controller) handleExpired(task base.ITask, dispatcher base.ITaskDispatcher) {
	defer dispatcher.OnError(task)
	//dispatcher.OnReady(task)
	log.Printf("Received expiration for %02x", task.Index())
}

func (rc *R606Controller) setupWork(work *stratum.Work) {
	rc.taskMtx.Lock()
	defer rc.taskMtx.Unlock()
	rc.nTimeOffset = 0
	started := rc.currentPoolWork != nil
	rc.currentPoolWork = work
	rc.currentPoolTask = stratum.NewTask(rc.currentPoolWork, R606MidstateCount, true)
	taskWait := rc.maxTaskWait + time.Millisecond
	for _, task := range rc.tasks {
		task.Stop()
		task.Start(rc.OnExpired, taskWait)
	}
	if !started {
		rc.ITaskDispatcher.Start()
		go func() {
			for _, task := range rc.tasks {
				rc.sendQueue <- task
			}
		}()
	}
}

func (rc *R606Controller) readLoop() {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	//var task *protocol.Task
	var taskResponse *protocol.TaskResponse
	rb := protocol.NewResponseBlock()
	readTicker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-rc.quitRead:
			readTicker.Stop()
			rc.waiter.Done()
			return
		case <-readTicker.C:
			buf.Reset()
			rb.Count = 0
			if err := rc.ReadTimeout(buf, R606ReadTimeout); err != nil {
				log.Println("Error reading response block:", err)
				continue
			}
			if err := rb.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
				log.Println("Error decoding response block:", err)
				continue
			}
			for i := 0; i < rb.Count; i++ {
				taskResponse = rb.Responses[i]
				if taskResponse.IsBusyWork() {
					continue
				}
				//task = rc.tasks[taskResponse.JobId]
				//rc.OnReceived(task)
				log.Printf("Received response for %02x  %02x %02x", taskResponse.JobId, taskResponse.Midstate,
					taskResponse.Nonce)
			}
		}
	}
}

func (rc *R606Controller) writeLoop() {
	var work *stratum.Work
	var next base.ITask
	var ok bool
	workChan := rc.WorkChannel()
	writeTicker := time.NewTicker(rc.fullscanDuration)
	for {
		select {
		case <-rc.quitRead:
			writeTicker.Stop()
			rc.waiter.Done()
			return
		case <-writeTicker.C:
			select {
			case next, ok = <-rc.sendQueue:
				if ok {
					log.Printf("Sending %02x", next.Index())
					rc.OnSend(next)
				}
			}
		case work = <-workChan:
			rc.setupWork(work)
		}
	}
}
