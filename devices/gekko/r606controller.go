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
	R606BaudDiv          = 1
	R606MinFrequency     = 200
	R606MaxFrequency     = 1200
	R606DefaultFrequency = 900
	R606NumCores         = 114
	R606NumChips         = 12
	R606MaxJobId         = 0x7f
	R606MidstateCount    = 4
	R606MaxVerifyTasks   = R606MidstateCount * R606MidstateCount * R606MaxJobId
	R606WaitFactor       = 0.5
)

type R606Controller struct {
	base.IController
	lastReset        time.Time
	frequency        float64
	chipCount        int
	tasks            []*protocol.Task
	quitMainLoop     chan struct{}
	quitPrepareLoop  chan struct{}
	quitVerifyLoop   chan struct{}
	waiter           sync.WaitGroup
	prepareQueue     chan *protocol.Task
	sendQueue        chan *protocol.Task
	resultsQueue     chan *base.TaskResult
	verifyQueue      chan *base.TaskResult
	versions         *utils.Versions
	currentPoolWork  *stratum.Work
	currentPoolTask  *stratum.Task
	taskMtx          sync.Mutex
	fullscanDuration time.Duration
	maxTaskWait      time.Duration
	currentDiff      *big.Int
	poolDiff         uint32
}

func NewR606Controller(controller base.IController) *R606Controller {
	rc := &R606Controller{IController: controller, quitMainLoop: make(chan struct{}),
		quitPrepareLoop: make(chan struct{}), quitVerifyLoop: make(chan struct{}),
		frequency: R606DefaultFrequency, currentDiff: big.NewInt(0)}
	rc.allocateTasks()
	return rc
}

func (rc *R606Controller) allocateTasks() {
	var task *protocol.Task
	var j byte
	rc.tasks = make([]*protocol.Task, R606MaxJobId)
	for j = 0; j < R606MaxJobId; j++ {
		task = protocol.NewTask(j, R606MidstateCount)
		task.SetBusyWork()
		rc.tasks[j] = task
	}
}

func (rc *R606Controller) Close() {
	waitForReader := rc.quitMainLoop != nil
	if rc.quitMainLoop != nil {
		close(rc.quitMainLoop)
	}
	if rc.quitPrepareLoop != nil {
		close(rc.quitPrepareLoop)
	}
	if rc.quitVerifyLoop != nil {
		close(rc.quitVerifyLoop)
	}
	if rc.sendQueue != nil {
		close(rc.sendQueue)
	}
	if waitForReader {
		rc.waiter.Wait()
	}
	rc.IController.Close()
}

func (rc *R606Controller) Reset() error {
	log.Println("Resetting", rc.LongString())
	if err := rc.performReset(); err != nil {
		go rc.Exit()
		return err
	}
	if err := rc.countChips(); err != nil {
		go rc.Exit()
		return err
	} else {
		log.Println(rc.LongString(), "found", rc.chipCount, "chips")
	}
	if err := rc.sendChainInactive(); err != nil {
		return err
	}
	if err := rc.setBaud(); err != nil {
		return err
	}
	for i := 0; i < rc.chipCount; i++ {
		if err := rc.setFrequency(R606DefaultFrequency, byte(i)); err != nil {
			return err
		}
	}
	rc.setTiming()
	if err := rc.initializeTasks(); err != nil {
		go rc.Exit()
		return err
	}
	log.Println(rc.LongString(), "reset")
	return nil
}

func (rc *R606Controller) performReset() error {
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

func (rc *R606Controller) countChips() error {
	var buf bytes.Buffer
	cc := protocol.NewCountChips()
	data, _ := cc.MarshalBinary()
	if err := rc.Write(data); err != nil {
		return err
	}
	time.Sleep(10 * time.Millisecond)
	if err := rc.ReadTimeout(&buf, 100*time.Millisecond); err != nil {
		return err
	} else {
		ccr := protocol.NewCountChipsResponse()
		if err := ccr.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
			return err
		} else {
			rc.chipCount = len(ccr.Chips)
			if rc.chipCount != R606NumChips {
				return fmt.Errorf("found %d chips instead of %d", rc.chipCount, R606NumChips)
			}
			return nil
		}
	}
}

func (rc *R606Controller) sendChainInactive() error {
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

func (rc *R606Controller) setBaud() error {
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

func (rc *R606Controller) setFrequency(frequency float64, asicId byte) error {
	sf := protocol.NewSetFrequency(R606MinFrequency, R606MaxFrequency, frequency, asicId)
	if sf.Frequency != rc.frequency {
		if data, err := sf.MarshalBinary(); err != nil {
			return err
		} else {
			if err := rc.Write(data); err != nil {
				return err
			}
			rc.setTiming()
		}
	}
	return nil
}

func (rc *R606Controller) setTiming() {
	hashRate := float64(rc.chipCount) * rc.frequency * R606NumCores * 1000000.0
	fullScanMicroSeconds := 1000000.0 * (float64(0xffffffff) / hashRate)
	rc.fullscanDuration = time.Duration(fullScanMicroSeconds*1000.0) * time.Nanosecond
	rc.maxTaskWait = time.Duration((R606WaitFactor * 4) * float64(rc.fullscanDuration))
	minTaskWait := 1 * time.Microsecond
	maxTaskWait := 3 * rc.fullscanDuration
	if rc.maxTaskWait < minTaskWait {
		rc.maxTaskWait = minTaskWait
	}
	if rc.maxTaskWait > maxTaskWait {
		rc.maxTaskWait = maxTaskWait
	}
	log.Println("Hashrate", utils.HashRate(hashRate))
	log.Println("Full scan time:", rc.fullscanDuration)
	log.Println("Max task wait:", rc.maxTaskWait)
}

func (rc *R606Controller) initializeTasks() error {
	rc.sendQueue = make(chan *protocol.Task, R606MaxJobId)
	rc.prepareQueue = make(chan *protocol.Task, R606MaxJobId)
	rc.resultsQueue = make(chan *base.TaskResult, R606MaxVerifyTasks)
	rc.verifyQueue = make(chan *base.TaskResult, R606MaxVerifyTasks)
	for i := 0; i < R606MaxVerifyTasks; i++ {
		rc.resultsQueue <- base.NewTaskResult()
	}
	rc.waiter.Add(3)
	go rc.prepareLoop()
	go rc.verifyLoop()
	go rc.mainLoop()
	return nil
}

func (rc *R606Controller) startWork() {
	for _, task := range rc.tasks {
		rc.sendQueue <- task
	}
}

func (rc *R606Controller) setupWork(work *stratum.Work) {
	rc.versions = utils.NewVersions(work.Version, work.VersionRollingMask, R606MidstateCount)
	rc.taskMtx.Lock()
	defer rc.taskMtx.Unlock()
	started := rc.currentPoolWork != nil
	rc.currentPoolWork = work
	rc.currentPoolTask = stratum.NewTask(rc.currentPoolWork, R606MidstateCount, false, rc.versions)
	utils.CalculateDifficulty(big.NewInt(int64(work.Difficulty)), rc.currentDiff)
	if !started {
		go rc.startWork()
	}
}

func (rc *R606Controller) prepareLoop() {
	var task *protocol.Task
	for {
		select {
		case <-rc.quitPrepareLoop:
			rc.waiter.Done()
			rc.quitVerifyLoop = nil
			return
		case task = <-rc.prepareQueue:
			if task.Index()%R606MidstateCount != 0 {
				continue
			}
			rc.taskMtx.Lock()
			rc.currentPoolWork.SetExtraNonce2(rc.currentPoolWork.ExtraNonce2 + 1)
			rc.currentPoolTask = stratum.NewTask(rc.currentPoolWork, R606MidstateCount, true, rc.versions)
			task.Update(rc.currentPoolTask)
			rc.taskMtx.Unlock()
			rc.sendQueue <- task
		}
	}
}

func (rc *R606Controller) mainLoop() {
	var midstate, index int
	var work *stratum.Work
	var currentTask *protocol.Task
	var ok bool
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	var nextResult *base.TaskResult
	var taskResponse *protocol.TaskResponse
	rb := protocol.NewResponseBlock()
	readTicker := time.NewTicker(2 * time.Millisecond)
	writeTicker := time.NewTicker(rc.fullscanDuration)
	expireTicker := time.NewTicker(rc.maxTaskWait)
	expireChan := make(chan *protocol.Task, R606MaxJobId)
	workChan := rc.WorkChannel()
	for {
		select {
		case <-rc.quitMainLoop:
			readTicker.Stop()
			writeTicker.Stop()
			expireTicker.Stop()
			rc.waiter.Done()
			rc.quitMainLoop = nil
			return
		case <-readTicker.C:
			buf.Reset()
			rb.Count = 0
			if err := rc.ReadTimeout(buf, time.Millisecond); err != nil {
				log.Println("Error reading response block:", err)
				rc.waiter.Done()
				go rc.Exit()
				return
			}
			if err := rb.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
				log.Println("Error decoding response block:", err)
				continue
			}
			for i := 0; i < rb.Count; i++ {
				taskResponse = rb.Responses[i]
				if taskResponse.BusyResponse() {
					continue
				}
				midstate = taskResponse.JobId % R606MidstateCount
				if midstate != 0 {
					index = taskResponse.JobId - midstate
				} else {
					index = taskResponse.JobId
				}
				if index > R606MaxJobId || index < 0 {
					continue
				}
				nextResult = <-rc.resultsQueue
				rc.tasks[index].UpdateResult(nextResult, taskResponse.Nonce, midstate)
				rc.verifyQueue <- nextResult
			}
		case <-writeTicker.C:
			select {
			case currentTask, ok = <-rc.sendQueue:
				if ok {
					if data, err := currentTask.MarshalBinary(); err != nil {
						panic(err)
					} else {
						if err = rc.Write(data); err != nil {
							log.Println("USB write error:", err)
							rc.waiter.Done()
							go rc.Exit()
							return
						}
					}
					if currentTask.IsBusyWork() {
						rc.prepareQueue <- currentTask
					} else {
						expireChan <- currentTask
					}
				}
			default:
			}
		case <-expireTicker.C:
			select {
			case currentTask, ok = <-expireChan:
				if ok {
					rc.prepareQueue <- currentTask
				}
			default:
			}
		case work = <-workChan:
			rc.setupWork(work)
		}
	}
}

func (rc *R606Controller) verifyLoop() {
	var verifyTask *base.TaskResult
	var resultDiff big.Int
	var diff utils.Difficulty
	var hashBig big.Int
	var match bool
	var matches, nonMatches uint64
	var submitVersion uint32
	for {
		select {
		case <-rc.quitVerifyLoop:
			rc.waiter.Done()
			rc.quitVerifyLoop = nil
			return
		case verifyTask = <-rc.verifyQueue:
			hash := verifyTask.CalculateHash()
			if hash[31] == 0 && hash[30] == 0 && hash[29] == 0 && hash[28] == 0 {
				matches += 1
				utils.HashToBig(hash, &hashBig)
				match = hashBig.Cmp(rc.currentDiff) <= 0
				if match {
					utils.CalculateDifficulty(&hashBig, &resultDiff)
					diff = utils.Difficulty(resultDiff.Int64())
					rc.taskMtx.Lock()
					if rc.currentPoolWork.VersionRolling {
						submitVersion = verifyTask.Version & ^rc.currentPoolWork.Version
					} else {
						submitVersion = 0
					}
					rc.taskMtx.Unlock()
					rc.currentPoolWork.Pool.SubmitChan <- protocol2.NewSubmit(
						verifyTask.JobId, verifyTask.ExtraNonce2, verifyTask.NTime, verifyTask.Nonce, submitVersion,
					)
					log.Printf("Extra nonce2 %016x nonce %08x version %08x diff %s",
						verifyTask.ExtraNonce2, verifyTask.Nonce, verifyTask.Version, diff)
				} else {
					if matches%0xff == 0 {
						log.Println(matches, "matches,", nonMatches, "non-matches")
					}
				}
			} else {
				nonMatches += 1
			}
			//log.Printf("Received nonce %02x, version %02x, ntime %02x, extraNonce %x for job id %s",
			//	verifyTask.Nonce, verifyTask.Version, verifyTask.NTime, verifyTask.ExtraNonce2, verifyTask.JobId)
			rc.resultsQueue <- verifyTask
		}
	}
}
