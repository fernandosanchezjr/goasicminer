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
}

func NewR606Controller(controller base.IController) *R606Controller {
	rc := &R606Controller{IController: controller, quit: make(chan struct{}), frequency: 0.0,
		currentDiff: big.NewInt(0)}
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
	rc.shuttingDown = true
	waitForLoops := rc.quit != nil
	if rc.quit != nil {
		close(rc.quit)
	}
	if rc.verifyQueue != nil {
		close(rc.verifyQueue)
	}
	if waitForLoops {
		rc.waiter.Wait()
		rc.quit = nil
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
	if err := rc.setFrequency(R606DefaultFrequency); err != nil {
		return err
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

func (rc *R606Controller) setFrequency(frequency float64) error {
	if frequency < R606MinFrequency {
		frequency = R606MinFrequency
	} else if frequency > R606MaxFrequency {
		frequency = R606MaxFrequency
	}
	if rc.frequency != frequency {
		for i := 0; i < rc.chipCount; i++ {
			if err := rc.setChipFrequency(R606DefaultFrequency, i); err != nil {
				return err
			}
		}
		rc.frequency = frequency
	}
	return nil
}

func (rc *R606Controller) setChipFrequency(frequency float64, chipId int) error {
	sf := protocol.NewSetFrequency(frequency, rc.chipCount, chipId)
	if data, err := sf.MarshalBinary(); err != nil {
		return err
	} else {
		if err := rc.Write(data); err != nil {
			return err
		}
		buf := bytes.NewBuffer(make([]byte, 0, 2048))
		if err := rc.ReadTimeout(buf, 50*time.Millisecond); err != nil {
			return err
		}
	}
	return nil
}

func (rc *R606Controller) setTiming() {
	var hashRate utils.HashRate
	hashRate, rc.fullscanDuration, rc.maxTaskWait = protocol.Timing(rc.chipCount, rc.frequency, R606NumCores,
		R606WaitFactor)
	log.Println("Hashrate", hashRate)
	log.Println("Full scan time:", rc.fullscanDuration)
	log.Println("Max task wait:", rc.maxTaskWait)
}

func (rc *R606Controller) initializeTasks() error {
	rc.verifyQueue = make(chan *base.TaskResult, R606MaxVerifyTasks)
	rc.waiter.Add(3)
	go rc.verifyLoop()
	go rc.readLoop()
	go rc.writeLoop()
	return nil
}

func (rc *R606Controller) loopRecover(loopName string) {
	if err := recover(); err != nil {
		log.Printf("Error in %s loop: %s", loopName, err)
		rc.waiter.Done()
		if !rc.shuttingDown {
			go rc.Exit()
		}
	}
}

func (rc *R606Controller) readLoop() {
	defer rc.loopRecover("read")
	var missingLoops uint32
	var midstate, index int
	buf := bytes.NewBuffer(make([]byte, 0, 2048))
	var nextResult *base.TaskResult
	var taskResponse *protocol.TaskResponse
	rb := protocol.NewResponseBlock()
	mainTicker := time.NewTicker(time.Millisecond)
	verifyTasks := make([]*base.TaskResult, R606MaxVerifyTasks)
	for i := 0; i < R606MaxVerifyTasks; i++ {
		verifyTasks[i] = base.NewTaskResult()
	}
	var verifyPos int
	for {
		select {
		case <-rc.quit:
			mainTicker.Stop()
			rc.waiter.Done()
			return
		case <-mainTicker.C:
			buf.Reset()
			if len(rb.ExtraData) > 0 {
				buf.Write(rb.ExtraData)
				rb.ExtraData = nil
			}
			rb.Count = 0
			if err := rc.Read(buf); err != nil {
				log.Println("Error reading response block:", err)
				rc.waiter.Done()
				go rc.Exit()
				return
			}
			if err := rb.UnmarshalBinary(protocol.Separator.Clean(buf.Bytes())); err != nil {
				log.Println("Error decoding response block:", err)
				continue
			}
			if rb.Count == 0 {
				missingLoops += 1
				if missingLoops >= 5000 {
					rc.waiter.Done()
					go rc.Exit()
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
				midstate = taskResponse.JobId % R606MidstateCount
				if midstate != 0 {
					index = taskResponse.JobId - midstate
				} else {
					index = taskResponse.JobId
				}
				if index > R606MaxJobId || index < 0 {
					continue
				}
				nextResult = verifyTasks[verifyPos]
				rc.tasks[index].UpdateResult(nextResult, taskResponse.Nonce, midstate)
				rc.verifyQueue <- nextResult
				verifyPos += 1
				if verifyPos >= R606MaxVerifyTasks {
					verifyPos = 0
				}
			}
		}
	}
}

func (rc *R606Controller) writeLoop() {
	defer rc.loopRecover("write")
	var task *stratum.Task
	var work *stratum.Work
	workChan := rc.WorkChannel()
	var versions *utils.Versions
	var versionMasks [R606MidstateCount]uint32
	versionTicker := time.NewTicker(15 * time.Second)
	var currentTask *protocol.Task
	mainTicker := time.NewTicker(rc.fullscanDuration)
	var nextPos uint32
	var warmedUp bool
	for {
		select {
		case <-rc.quit:
			mainTicker.Stop()
			versionTicker.Stop()
			rc.waiter.Done()
			return
		case <-mainTicker.C:
			if work == nil {
				continue
			}
			currentTask = rc.tasks[nextPos]
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
			if !warmedUp {
				nextPos += 1
			} else {
				nextPos += R606MidstateCount
			}
			if nextPos >= R606MaxJobId {
				warmedUp = true
				nextPos = 0
			}
			if warmedUp {
				currentTask = rc.tasks[nextPos]
				work.SetExtraNonce2(work.ExtraNonce2 + 1)
				task.Update(work, versionMasks[:])
				currentTask.Update(task)
			}
		case work = <-workChan:
			if versions == nil {
				versions = utils.NewVersions(work.Version, work.VersionRollingMask, R606MidstateCount)
				versions.Retrieve(versionMasks[:])
			}
			task = stratum.NewTask(R606MidstateCount, true)
			rc.submitChan = work.Pool.SubmitChan
			rc.poolVersion = work.Version
			rc.poolVersionRolling = work.VersionRolling
			utils.CalculateDifficulty(big.NewInt(int64(work.Difficulty)), rc.currentDiff)
		case <-versionTicker.C:
			if versions != nil {
				versions.Retrieve(versionMasks[:])
			}
		}
	}
}

func (rc *R606Controller) verifyLoop() {
	defer rc.loopRecover("verify")
	var verifyTask *base.TaskResult
	var resultDiff big.Int
	var diff utils.Difficulty
	var hashBig big.Int
	var match bool
	var submitVersion uint32
	var maxDiff utils.Difficulty
	for {
		select {
		case <-rc.quit:
			rc.waiter.Done()
			return
		case verifyTask = <-rc.verifyQueue:
			if verifyTask == nil {
				continue
			}
			hash := verifyTask.CalculateHash()
			if hash[31] == 0 && hash[30] == 0 && hash[29] == 0 && hash[28] == 0 {
				utils.HashToBig(hash, &hashBig)
				match = hashBig.Cmp(rc.currentDiff) <= 0
				if match {
					utils.CalculateDifficulty(&hashBig, &resultDiff)
					diff = utils.Difficulty(resultDiff.Int64())
					if rc.poolVersionRolling {
						submitVersion = verifyTask.Version & ^rc.poolVersion
					} else {
						submitVersion = 0
					}
					rc.submitChan <- protocol2.NewSubmit(
						verifyTask.JobId, verifyTask.ExtraNonce2, verifyTask.NTime, verifyTask.Nonce, submitVersion,
					)
					log.Printf("Job id %s extra nonce2 %016x ntime %08x nonce %08x version %08x diff %s",
						verifyTask.JobId, verifyTask.ExtraNonce2, verifyTask.NTime, verifyTask.Nonce,
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
