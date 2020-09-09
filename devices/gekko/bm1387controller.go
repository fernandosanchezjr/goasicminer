package gekko

import (
	"bytes"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/devices/base"
	"github.com/fernandosanchezjr/goasicminer/devices/gekko/protocol"
	"github.com/fernandosanchezjr/goasicminer/stratum"
	protocol2 "github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"log"
	"math/big"
	"sync"
	"time"
)

const (
	BM1387BaudDiv        = 1
	BM1387NumCores       = 114
	BM1387MaxJobId       = 0x7f
	BM1387MidstateCount  = 4
	BM1387MaxVerifyTasks = BM1387MidstateCount * BM1387MidstateCount * BM1387MaxJobId
	BM1387WaitFactor     = 0.5
)

type BM1387Controller struct {
	base.IController
	lastReset          time.Time
	frequency          float64
	chipCount          int
	tasks              []*protocol.Task
	quit               chan struct{}
	waiter             sync.WaitGroup
	verifyQueue        chan *base.TaskResult
	fullscanDuration   time.Duration
	maxTaskWait        time.Duration
	currentDiff        *big.Int
	poolVersion        uint32
	poolVersionRolling bool
	shuttingDown       bool
	submitChan         chan *protocol2.Submit
	minFrequency       float64
	maxFrequency       float64
	defaultFrequency   float64
	targetChips        int
}

func NewBM1387Controller(
	controller base.IController,
	minFrequency float64,
	maxFrequency float64,
	defaultFrequency float64,
	targetChips int,
) *BM1387Controller {
	rc := &BM1387Controller{IController: controller, quit: make(chan struct{}), frequency: 0.0,
		currentDiff: big.NewInt(0), minFrequency: minFrequency, maxFrequency: maxFrequency,
		defaultFrequency: defaultFrequency, targetChips: targetChips}
	rc.allocateTasks()
	return rc
}

func (bm *BM1387Controller) allocateTasks() {
	var task *protocol.Task
	var j byte
	bm.tasks = make([]*protocol.Task, BM1387MaxJobId)
	for j = 0; j < BM1387MaxJobId; j++ {
		task = protocol.NewTask(j, BM1387MidstateCount)
		task.SetBusyWork()
		bm.tasks[j] = task
	}
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
	log.Println("Resetting", bm.LongString())
	if err := bm.performReset(); err != nil {
		go bm.Exit()
		return err
	}
	if err := bm.countChips(); err != nil {
		go bm.Exit()
		return err
	} else {
		log.Println(bm.LongString(), "found", bm.chipCount, "chips")
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
	bm.setTiming()
	if err := bm.initializeTasks(); err != nil {
		go bm.Exit()
		return err
	}
	log.Println(bm.LongString(), "reset")
	return nil
}

func (bm *BM1387Controller) performReset() error {
	device := bm.USBDevice()
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

func (bm *BM1387Controller) countChips() error {
	var buf bytes.Buffer
	cc := protocol.NewCountChips()
	data, _ := cc.MarshalBinary()
	if err := bm.Write(data); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if err := bm.ReadTimeout(&buf, 100*time.Millisecond); err != nil {
		return err
	} else {
		ccr := protocol.NewCountChipsResponse()
		if err := ccr.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
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
		if err := bm.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
		if err := bm.Write(data); err != nil {
			return err
		}
		time.Sleep(5 * time.Millisecond)
		if err := bm.Write(data); err != nil {
			return err
		}
	}
	for i := 0; i < bm.chipCount; i++ {
		cic.SetCurrentChip(i)
		if data, err := cic.MarshalBinary(); err != nil {
			return err
		} else {
			time.Sleep(5 * time.Millisecond)
			if err := bm.Write(data); err != nil {
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
		if err := bm.Write(data); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Set baud
	if _, err := bm.USBDevice().Control(64, 3, 0x02, 0, nil); err != nil {
		return err
	}
	if data, err := sbb.MarshalBinary(); err != nil {
		return err
	} else {
		if err := bm.Write(data); err != nil {
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
	sf := protocol.NewSetFrequency(frequency, bm.chipCount, chipId)
	if data, err := sf.MarshalBinary(); err != nil {
		return err
	} else {
		if err := bm.Write(data); err != nil {
			return err
		}
		buf := bytes.NewBuffer(make([]byte, 0, 2048))
		if err := bm.ReadTimeout(buf, 50*time.Millisecond); err != nil {
			return err
		}
	}
	return nil
}

func (bm *BM1387Controller) setTiming() {
	var hashRate utils.HashRate
	hashRate, bm.fullscanDuration, bm.maxTaskWait = protocol.Timing(bm.chipCount, bm.frequency, BM1387NumCores,
		BM1387WaitFactor)
	log.Println("Hashrate", hashRate)
	log.Println("Full scan time:", bm.fullscanDuration)
	log.Println("Max task wait:", bm.maxTaskWait)
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
		log.Printf("Error in %s loop: %s", loopName, err)
		bm.waiter.Done()
		if !bm.shuttingDown {
			go bm.Exit()
		}
	}
}

func (bm *BM1387Controller) readLoop() {
	defer bm.loopRecover("read")
	var missingLoops uint32
	var midstate, index int
	buf := bytes.NewBuffer(make([]byte, 0, 2048))
	var nextResult *base.TaskResult
	var taskResponse *protocol.TaskResponse
	rb := protocol.NewResponseBlock()
	mainTicker := time.NewTicker(time.Millisecond)
	verifyTasks := make([]*base.TaskResult, BM1387MaxVerifyTasks)
	for i := 0; i < BM1387MaxVerifyTasks; i++ {
		verifyTasks[i] = base.NewTaskResult()
	}
	var verifyPos int
	for {
		select {
		case <-bm.quit:
			mainTicker.Stop()
			bm.waiter.Done()
			return
		case <-mainTicker.C:
			buf.Reset()
			if len(rb.ExtraData) > 0 {
				buf.Write(rb.ExtraData)
				rb.ExtraData = nil
			}
			rb.Count = 0
			if err := bm.Read(buf); err != nil {
				log.Println("Error reading response block:", err)
				bm.waiter.Done()
				go bm.Exit()
				return
			}
			if err := rb.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
				log.Println("Error decoding response block:", err)
				continue
			}
			if rb.Count == 0 {
				missingLoops += 1
				if missingLoops >= 1000 {
					bm.waiter.Done()
					go bm.Exit()
					return
				}
				continue
			}
			if rb.Count >= len(rb.Responses) {
				rb.ExtraData = nil
				continue
			}
			missingLoops = 0
			for i := 0; i < rb.Count; i++ {
				taskResponse = rb.Responses[i]
				if taskResponse.BusyResponse() {
					continue
				}
				midstate = taskResponse.JobId % BM1387MidstateCount
				if midstate != 0 {
					index = taskResponse.JobId - midstate
				} else {
					index = taskResponse.JobId
				}
				if index > BM1387MaxJobId || index < 0 {
					continue
				}
				nextResult = verifyTasks[verifyPos]
				bm.tasks[index].UpdateResult(nextResult, taskResponse.Nonce, midstate)
				bm.verifyQueue <- nextResult
				verifyPos += 1
				if verifyPos >= BM1387MaxVerifyTasks {
					verifyPos = 0
				}
			}
		}
	}
}

func (bm *BM1387Controller) writeLoop() {
	defer bm.loopRecover("write")
	var task *stratum.Task
	var work *stratum.Work
	workChan := bm.WorkChannel()
	var versionMasks [BM1387MidstateCount]uint32
	var currentTask *protocol.Task
	mainTicker := time.NewTicker(bm.fullscanDuration)
	var nextPos uint32
	var warmedUp bool
	for {
		select {
		case <-bm.quit:
			mainTicker.Stop()
			bm.waiter.Done()
			return
		case <-mainTicker.C:
			if work == nil {
				continue
			}
			currentTask = bm.tasks[nextPos]
			if data, err := currentTask.MarshalBinary(); err != nil {
				panic(err)
			} else {
				if err = bm.Write(data); err != nil {
					log.Println("USB write error:", err)
					bm.waiter.Done()
					go bm.Exit()
					return
				}
			}
			if !warmedUp {
				nextPos += 1
			} else {
				nextPos += BM1387MidstateCount
			}
			if nextPos >= BM1387MaxJobId {
				warmedUp = true
				nextPos = 0
			}
			if warmedUp {
				currentTask = bm.tasks[nextPos]
				work.SetExtraNonce2(work.ExtraNonce2 + 1)
				work.VersionsSource.Retrieve(versionMasks[:])
				task.Update(work, versionMasks[:])
				currentTask.Update(task)
			}
		case work = <-workChan:
			task = stratum.NewTask(BM1387MidstateCount, true)
			bm.submitChan = work.Pool.SubmitChan
			bm.poolVersion = work.Version
			bm.poolVersionRolling = work.VersionRolling
			utils.CalculateDifficulty(big.NewInt(int64(work.Difficulty)), bm.currentDiff)
			if warmedUp {
				currentTask = bm.tasks[nextPos]
				work.SetExtraNonce2(work.ExtraNonce2 + 1)
				work.VersionsSource.Retrieve(versionMasks[:])
				task.Update(work, versionMasks[:])
				currentTask.Update(task)
			}
		}
	}
}

func (bm *BM1387Controller) verifyLoop() {
	defer bm.loopRecover("verify")
	var verifyTask *base.TaskResult
	var resultDiff big.Int
	var diff utils.Difficulty
	var hashBig big.Int
	var match bool
	var submitVersion uint32
	var maxDiff utils.Difficulty
	for {
		select {
		case <-bm.quit:
			bm.waiter.Done()
			return
		case verifyTask = <-bm.verifyQueue:
			if verifyTask == nil {
				continue
			}
			hash := verifyTask.CalculateHash()
			if hash[31] == 0 && hash[30] == 0 && hash[29] == 0 && hash[28] == 0 {
				utils.HashToBig(hash, &hashBig)
				match = hashBig.Cmp(bm.currentDiff) <= 0
				if match {
					if bm.poolVersionRolling {
						submitVersion = verifyTask.Version & ^bm.poolVersion
					} else {
						submitVersion = 0
					}
					bm.submitChan <- protocol2.NewSubmit(
						verifyTask.JobId, verifyTask.ExtraNonce2, verifyTask.NTime, verifyTask.Nonce, submitVersion,
					)
					utils.CalculateDifficulty(&hashBig, &resultDiff)
					diff = utils.Difficulty(resultDiff.Int64())
					log.Printf("%s job %s extra nonce2 %016x ntime %08x nonce %08x version %08x diff %s",
						bm, verifyTask.JobId, verifyTask.ExtraNonce2, verifyTask.NTime, verifyTask.Nonce,
						verifyTask.Version, diff)
					if diff > maxDiff {
						maxDiff = diff
						log.Println("Best share:", maxDiff)
					}
				}
			}
		}
	}
}
