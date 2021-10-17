package gekko

import (
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko/protocol"
	"github.com/fernandosanchezjr/goasicminer/generators"
	"github.com/fernandosanchezjr/goasicminer/node"
	protocol2 "github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"github.com/ziutek/ftdi"
	"math/big"
	"strings"
	"sync"
	"time"
)

const (
	BM1387BaudDiv         = 1
	BM1387InitialBaudRate = 115200
	BM1387BaudRate        = 375000
	BM1387NumCores        = 114
	BM1387MaxJobId        = 0x7f
	BM1387MidstateCount   = 4
	BM1387MaxVerifyTasks  = BM1387MidstateCount * BM1387MidstateCount * BM1387MaxJobId
	BM1387WaitFactor      = 0.5
)

type BM1387Controller struct {
	base.IController
	frequency        float64
	chipCount        int
	quit             chan struct{}
	waiter           sync.WaitGroup
	verifyQueue      chan *base.TaskResult
	timeout          time.Duration
	fullscanDuration time.Duration
	maxTaskWait      time.Duration
	currentDiff      *big.Int
	targetDiff       *big.Int
	poolVersion      utils.Version
	shuttingDown     bool
	submitChan       chan *protocol2.Submit
	minFrequency     float64
	maxFrequency     float64
	defaultFrequency float64
	targetChips      int
	warmupWritten    bool
	warmupRead       bool
	work             *node.Work
	lastRead         time.Time
	readTicker       *time.Ticker
	writeTicker      *time.Ticker
	taskResultPool   *base.TaskResultPool
	pendingTaskPool  *protocol.TaskPool
}

func NewBM1387Controller(
	controller base.IController,
	minFrequency float64,
	maxFrequency float64,
	defaultFrequency float64,
	targetChips int,
	timeout time.Duration,
) *BM1387Controller {
	rc := &BM1387Controller{IController: controller, quit: make(chan struct{}), frequency: 0.0,
		currentDiff: big.NewInt(0), targetDiff: big.NewInt(0), minFrequency: minFrequency, maxFrequency: maxFrequency,
		defaultFrequency: defaultFrequency, targetChips: targetChips, timeout: timeout,
		taskResultPool:  base.NewTaskResultPool(BM1387MaxVerifyTasks),
		pendingTaskPool: protocol.NewTaskPool(BM1387MaxJobId, BM1387MidstateCount),
	}
	return rc
}

func (bm *BM1387Controller) Close() {
	bm.shuttingDown = true
	waitForLoops := bm.quit != nil
	if bm.quit != nil {
		close(bm.quit)
	}
	if bm.verifyQueue != nil {
		close(bm.verifyQueue)
	}
	if waitForLoops {
		bm.waiter.Wait()
		bm.quit = nil
	}
	bm.IController.Close()
}

func (bm *BM1387Controller) Reset() error {
	defer bm.loopRecover("init")
	if err := bm.performReset(); err != nil {
		go bm.Exit()
		return err
	}
	if err := bm.countChips(); err != nil {
		go bm.Exit()
		return err
	}
	if err := bm.sendChainInactive(); err != nil {
		return err
	}
	if err := bm.setBaud(); err != nil {
		return err
	}
	if err := bm.setFrequency(bm.defaultFrequency); err != nil {
		return err
	}
	if err := bm.setTiming(); err != nil {
		return err
	}
	if err := bm.initializeTasks(); err != nil {
		go bm.Exit()
		return err
	}
	return nil
}

func (bm *BM1387Controller) performReset() error {
	device := bm.Device()
	if err := device.Reset(); err != nil {
		return err
	}
	if err := device.SetLineProperties2(ftdi.DataBits8, ftdi.StopBits1, ftdi.ParityNone, ftdi.BreakOff); err != nil {
		return err
	}
	if err := device.SetBaudrate(BM1387InitialBaudRate); err != nil {
		return err
	}
	if err := device.SetFlowControl(ftdi.FlowCtrlDisable); err != nil {
		return err
	}
	if err := device.PurgeWriteBuffer(); err != nil {
		return err
	}
	if err := device.PurgeReadBuffer(); err != nil {
		return err
	}
	if err := device.SetBitmode(0xf2, ftdi.ModeCBUS); err != nil {
		return err
	}
	time.Sleep(30 * time.Millisecond)
	if err := device.SetBitmode(0xf0, ftdi.ModeCBUS); err != nil {
		return err
	}
	time.Sleep(30 * time.Millisecond)
	if err := device.SetBitmode(0xf2, ftdi.ModeCBUS); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (bm *BM1387Controller) countChips() error {
	buf, err := bm.AllocateReadBuffer()
	if err != nil {
		return err
	}
	cc := protocol.NewCountChips()
	data, _ := cc.MarshalBinary()
	if _, err := bm.Write(data); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if read, err := bm.Read(buf); err != nil {
		return err
	} else {
		ccr := protocol.NewCountChipsResponse()
		if err := ccr.UnmarshalBinary(buf[:read]); err != nil {
			return err
		} else {
			bm.chipCount = len(ccr.Chips)
			if bm.chipCount != bm.targetChips {
				return fmt.Errorf("found %d chips instead of %d", bm.chipCount, bm.targetChips)
			}
			return nil
		}
	}
}

func (bm *BM1387Controller) sendChainInactive() error {
	ci := protocol.NewChainInactive()
	cic := protocol.NewChainInactiveChip(bm.chipCount)
	if data, err := ci.MarshalBinary(); err != nil {
		return err
	} else {
		if _, err := bm.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
		if _, err := bm.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
		if _, err := bm.Write(data); err != nil {
			return err
		}
	}
	for i := 0; i < bm.chipCount; i++ {
		cic.SetCurrentChip(i)
		if data, err := cic.MarshalBinary(); err != nil {
			return err
		} else {
			time.Sleep(5 * time.Millisecond)
			if _, err := bm.Write(data); err != nil {
				return err
			}
		}
	}
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (bm *BM1387Controller) setBaud() error {
	sba := protocol.NewSetBaudA(BM1387BaudDiv)
	sbb := protocol.NewSetBaudGateBlockMessage(BM1387BaudDiv)
	if data, err := sba.MarshalBinary(); err != nil {
		return err
	} else {
		if _, err := bm.Write(data); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	//Set baud
	device := bm.Device()
	if err := device.SetBaudrate(BM1387BaudRate); err != nil {
		return err
	}
	if data, err := sbb.MarshalBinary(); err != nil {
		return err
	} else {
		if _, err := bm.Write(data); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (bm *BM1387Controller) setFrequency(frequency float64) error {
	if frequency < bm.minFrequency {
		frequency = bm.minFrequency
	} else if frequency > bm.maxFrequency {
		frequency = bm.maxFrequency
	}
	if bm.frequency != frequency {
		for i := 0; i < bm.chipCount; i++ {
			if err := bm.setChipFrequency(frequency, i); err != nil {
				return err
			}
		}
		bm.frequency = frequency
	}
	return nil
}

func (bm *BM1387Controller) setChipFrequency(frequency float64, chipId int) error {
	buf, err := bm.AllocateReadBuffer()
	if err != nil {
		return err
	}
	sf := protocol.NewSetFrequency(frequency, bm.chipCount, chipId)
	if data, err := sf.MarshalBinary(); err != nil {
		return err
	} else {
		if _, err := bm.Write(data); err != nil {
			return err
		}
		if _, err := bm.Read(buf); err != nil {
			return err
		}
	}
	return nil
}

func (bm *BM1387Controller) setTiming() error {
	var hashRate utils.HashRate
	hashRate, bm.fullscanDuration, bm.maxTaskWait = protocol.Timing(bm.chipCount, bm.frequency, BM1387NumCores,
		BM1387WaitFactor)
	if err := bm.Device().SetLatencyTimer(1); err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"serial":       bm.String(),
		"frequency":    bm.frequency,
		"hashRate":     hashRate,
		"fullScanTime": bm.fullscanDuration,
		"maxTaskWait":  bm.maxTaskWait,
	}).Infoln("Timing set up")
	return nil
}

func (bm *BM1387Controller) initializeTasks() error {
	bm.verifyQueue = make(chan *base.TaskResult, BM1387MaxVerifyTasks)
	bm.waiter.Add(3)
	go bm.verifyLoop()
	go bm.readLoop()
	go bm.writeLoop()
	return nil
}

func (bm *BM1387Controller) loopRecover(loopName string) {
	if err := recover(); err != nil {
		if !strings.Contains(fmt.Sprint(err), "send on closed channel") &&
			!strings.Contains(fmt.Sprint(err), "nil pointer dereference") {
			log.WithFields(log.Fields{
				"serial": bm.String(),
				"loop":   loopName,
				"error":  fmt.Errorf("%#v", err),
			}).Error("Loop error")
		}
		bm.waiter.Done()
		if !bm.shuttingDown {
			bm.Exit()
		}
	}
}

func closeTicker(ticker *time.Ticker) *time.Ticker {
	if ticker != nil {
		ticker.Stop()
	}
	return nil
}

func (bm *BM1387Controller) handlerExit() {
	bm.readTicker = closeTicker(bm.readTicker)
	bm.writeTicker = closeTicker(bm.writeTicker)
	bm.waiter.Done()
	bm.Exit()
}

func (bm *BM1387Controller) readLoop() {
	defer bm.loopRecover("read")
	buf, err := bm.AllocateReadBuffer()
	if err != nil {
		panic(err)
	}
	rb := protocol.NewResponseBlock()
	bm.readTicker = time.NewTicker(bm.fullscanDuration)
	var readTime time.Time
	var markRead bool
	for {
		select {
		case <-bm.quit:
			bm.handlerExit()
			return
		case readTime = <-bm.readTicker.C:
			markRead = false
			if !bm.warmupWritten {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			if read, died := bm.readResponseBlock(buf, rb); died {
				bm.handlerExit()
				return
			} else if !read {
				continue
			}
			markRead = bm.dispatchResponseValidation(rb)
			if markRead {
				bm.lastRead = readTime
			}
		}
	}
}

func (bm *BM1387Controller) readResponseBlock(buf []byte, rb *protocol.ResponseBlock) (read bool, died bool) {
	var readCount int
	var err error
	if bm.warmupRead && time.Since(bm.lastRead) > bm.timeout {
		return false, true
	}
	if readCount, err = bm.Read(buf); err != nil {
		log.WithFields(log.Fields{
			"serial": bm.String(),
			"error":  err.Error(),
		}).Error("Read error")
		return false, true
	}
	if err := rb.UnmarshalBinary(buf[:readCount]); err != nil {
		log.WithFields(log.Fields{
			"serial": bm.String(),
			"error":  err.Error(),
		}).Error("Error decoding response block")
		return false, false
	}
	return true, false
}

func (bm *BM1387Controller) dispatchResponseValidation(
	rb *protocol.ResponseBlock,
) bool {
	var read bool
	var index, midstate int
	var task *protocol.Task
	var taskResponse *protocol.TaskResponse
	for i := 0; i < rb.Count; i++ {
		taskResponse = rb.Responses[i]
		if taskResponse.BusyResponse() {
			read = true
			continue
		}
		bm.warmupRead = true
		midstate = taskResponse.JobId % BM1387MidstateCount
		if midstate != 0 {
			index = taskResponse.JobId - midstate
		} else {
			index = taskResponse.JobId
		}
		if index >= BM1387MaxJobId || index < 0 {
			continue
		}
		var nextResult = bm.taskResultPool.Next()
		task = bm.pendingTaskPool.GetTask(index)
		task.UpdateResult(nextResult, taskResponse.Nonce, midstate)
		bm.verifyQueue <- nextResult
		read = true
	}
	return read
}

func (bm *BM1387Controller) writeLoop() {
	defer bm.loopRecover("write")
	var generated *generators.Generated
	var generatorChan = bm.GetGenerator()
	var task = node.NewTask(BM1387MidstateCount, true)
	var workChan = bm.WorkChannel()
	var versionMasks [BM1387MidstateCount]utils.Version
	bm.writeTicker = time.NewTicker(bm.fullscanDuration)
	var steps = 1
	var currentTask, last = bm.pendingTaskPool.Next(steps)
	for {
		select {
		case <-bm.quit:
			bm.handlerExit()
			return
		case bm.work = <-workChan:
			continue
		case <-bm.writeTicker.C:
			if bm.work == nil {
				continue
			}
			if !bm.writeTask(currentTask) {
				bm.handlerExit()
				return
			}
			currentTask, last = bm.pendingTaskPool.Next(steps)
			if last {
				bm.warmupWritten = true
				steps = BM1387MidstateCount
			}
			if bm.warmupWritten {
				generated = <-generatorChan
				versionMasks[0] = generated.Version0
				versionMasks[1] = generated.Version1
				versionMasks[2] = generated.Version2
				versionMasks[3] = generated.Version3
				task.Update(generated.Work, versionMasks[:])
				currentTask.Update(task)
			}
		}
	}
}

func (bm *BM1387Controller) writeTask(currentTask *protocol.Task) (succeeded bool) {
	var written int
	var data []byte
	var err error
	if data, err = currentTask.MarshalBinary(); err != nil {
		log.WithFields(log.Fields{
			"serial": bm.String(),
			"error":  err.Error(),
		}).Error("Task marshalling error")
		bm.handlerExit()
		return false
	}
	if written, err = bm.Write(data); err != nil {
		log.WithFields(log.Fields{
			"serial": bm.String(),
			"error":  err.Error(),
		}).Error("Write error")
		bm.handlerExit()
		return false
	}
	if written != len(data) {
		log.WithFields(log.Fields{
			"serial": bm.String(),
			"error":  err.Error(),
		}).Error("Write incomplete")
		bm.handlerExit()
		return false
	}
	return true
}

func (bm *BM1387Controller) verifyLoop() {
	defer bm.loopRecover("verify")
	var task *base.TaskResult
	for {
		select {
		case <-bm.quit:
			bm.waiter.Done()
			return
		case task = <-bm.verifyQueue:
			task.Verify(bm.String())
		}
	}
}
