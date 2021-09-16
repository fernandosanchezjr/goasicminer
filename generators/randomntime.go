package generators

import (
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
	"time"
)

const (
	RandomNTimeBufferSize         = 64
	RandomNTimeGeneratedCacheSize = 8192
)

type RandomNTime struct {
	rng            *rand.Rand
	allVersions    []utils.Version
	nTime          utils.NTime
	minNTime       utils.NTime
	nTimeRI        *utils.RandomIndex
	quitChan       chan struct{}
	versionChan    chan *utils.VersionSource
	workChan       chan *node.Work
	generatedChan  chan *Generated
	knownNonceChan chan utils.Nonce64
	waiter         sync.WaitGroup
	progressChan   chan utils.Nonce64
	versionPos     int
	versionCount   int
	workId         uint64
}

func NewRandomNTime() *RandomNTime {
	var pb = &RandomNTime{
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		quitChan:       make(chan struct{}),
		versionChan:    make(chan *utils.VersionSource),
		workChan:       make(chan *node.Work),
		generatedChan:  make(chan *Generated, RandomNTimeBufferSize),
		knownNonceChan: make(chan utils.Nonce64, RandomNTimeGeneratedCacheSize),
		progressChan:   make(chan utils.Nonce64, RandomNTimeGeneratedCacheSize),
		allVersions:    utils.GetUsedVersions(),
	}
	pb.versionCount = len(pb.allVersions)
	pb.waiter.Add(1)
	go pb.generatorLoop()
	return pb
}

func (pb *RandomNTime) UpdateVersion(
	versionSource *utils.VersionSource,
) {
	pb.versionChan <- versionSource
}

func (pb *RandomNTime) UpdateWork(work *node.Work) {
	pb.workChan <- work
}

func (pb *RandomNTime) ExtraNonceFound(extraNonce utils.Nonce64) {
	pb.knownNonceChan <- extraNonce
}

func (pb *RandomNTime) Next(generated *Generated) bool {
	var versions utils.Versions
	var end bool
	generated.NTime = pb.nTime

	for i := 0; i < 4; i++ {
		if pb.versionPos >= pb.versionCount {
			pb.versionPos = 0
			end = true
		}
		versions[i] = pb.allVersions[pb.versionPos]
		pb.versionPos += 1
	}

	generated.Version0 = versions[0]
	generated.Version1 = versions[1]
	generated.Version2 = versions[2]
	generated.Version3 = versions[3]

	if end {
		pb.nTime = pb.minNTime + utils.NTime(pb.nTimeRI.Next(pb.rng))
		pb.versionPos = 0
	}
	return end
}

func (pb *RandomNTime) generatorLoop() {
	var pending, i int
	var work *node.Work
	var sent int
	for {
		select {
		case <-pb.quitChan:
			pb.waiter.Done()
			return
		case work = <-pb.workChan:
			pb.minNTime = work.MinNtime
			pb.versionPos = 0
			pb.workId = work.WorkId
			sent = 0
			pb.nTimeRI = utils.NewRandomIndex(int(work.Ntime - pb.minNTime + 300))
			pb.nTime = pb.minNTime + utils.NTime(pb.nTimeRI.Next(pb.rng))
		default:
			if work == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = RandomNTimeBufferSize - len(pb.generatedChan)
			if pending == 0 {
				continue
			}
			for i = 0; i < pending; i++ {
				var tmpGenerated = &Generated{}
				var end = pb.Next(tmpGenerated)
				pb.generatedChan <- tmpGenerated
				if end {
					work.Node.GenerateWorkAsync(pb.rng.Intn(work.TotalTransactions - 1))
				}
				sent += 4
			}
		}
	}
}

func (pb *RandomNTime) Close() {
	close(pb.quitChan)
	pb.waiter.Wait()
}

func (pb *RandomNTime) GeneratorChan() chan *Generated {
	return pb.generatedChan
}

func (pb *RandomNTime) ProgressChan() chan utils.Nonce64 {
	return pb.progressChan
}
