package generators

import (
	"github.com/fernandosanchezjr/goasicminer/analytics"
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
	"time"
)

const (
	UsedNTimesPacketSize         = 16
	UsedNTimesBufferSize         = 1024
	UsedNTimesGeneratedCacheSize = 8192
)

type PreviouslyUsedNTimes struct {
	rng            *rand.Rand
	allVersions    []utils.Version
	nTime          utils.NTime
	minnTime       utils.NTime
	quitChan       chan struct{}
	versionChan    chan *utils.VersionSource
	workChan       chan *node.Work
	generatedChan  chan *Generated
	knownNonceChan chan utils.Nonce64
	waiter         sync.WaitGroup
	progressChan   chan utils.Nonce64
	workId         uint64
	usedNTimes     *analytics.UsedNTimes
	nTimeReuse     int
}

func NewUsedNTimes() *PreviouslyUsedNTimes {
	var usedNTimes, loadErr = analytics.LoadRawUsedNtimes()
	if loadErr != nil {
		panic(loadErr)
	}
	var pb = &PreviouslyUsedNTimes{
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		quitChan:       make(chan struct{}),
		versionChan:    make(chan *utils.VersionSource),
		workChan:       make(chan *node.Work),
		generatedChan:  make(chan *Generated, UsedNTimesGeneratedCacheSize),
		knownNonceChan: make(chan utils.Nonce64, UsedNTimesGeneratedCacheSize),
		progressChan:   make(chan utils.Nonce64, UsedNTimesGeneratedCacheSize),
		allVersions:    utils.GetUsedVersions(),
		usedNTimes:     usedNTimes,
	}
	pb.waiter.Add(1)
	go pb.generatorLoop()
	return pb
}

func (pb *PreviouslyUsedNTimes) UpdateVersion(
	versionSource *utils.VersionSource,
) {
	pb.versionChan <- versionSource
}

func (pb *PreviouslyUsedNTimes) UpdateWork(work *node.Work) {
	pb.workChan <- work
}

func (pb *PreviouslyUsedNTimes) ExtraNonceFound(extraNonce utils.Nonce64) {
	pb.knownNonceChan <- extraNonce
}

func (pb *PreviouslyUsedNTimes) Next(generated *Generated, work *node.Work) {
	generated.Work = work.Clone()
	var versions utils.Versions
	var nTime utils.NTime

	nTime, versions = pb.usedNTimes.Next()
	generated.NTime = (pb.nTime & 0xffffff00) | nTime

	generated.Work.SetNtime(generated.NTime)
	generated.Version0 = versions[0]
	generated.Version1 = versions[1]
	generated.Version2 = versions[2]
	generated.Version3 = versions[3]

	return
}

func (pb *PreviouslyUsedNTimes) generatorLoop() {
	var pending, i int
	var work *node.Work
	var versionMask utils.Version
	var txCountRI *utils.RandomIndex
	var txCount int
	var sent int
	var reset bool
	for {
		select {
		case <-pb.quitChan:
			pb.waiter.Done()
			return
		case work = <-pb.workChan:
			if versionMask != work.Version {
				versionMask = work.Version
				pb.usedNTimes.FilterVersions(versionMask)
			}
			pb.nTime = work.Ntime
			pb.minnTime = work.MinNtime
			pb.workId = work.WorkId
			if txCountRI == nil || txCount != work.TotalTransactions {
				txCount = work.TotalTransactions
				txCountRI = utils.NewRandomIndex(txCount)
				txCountRI.Shuffle(pb.rng)
			}
			pb.usedNTimes.Reset()
			sent = 0
			reset = false
		default:
			if work == nil || txCountRI == nil {
				time.Sleep(time.Millisecond)
				continue
			}
			pending = UsedNTimesPacketSize - len(pb.generatedChan)
			if pending == 0 {
				continue
			}
			for i = 0; i < utils.Min(pending, UsedNTimesPacketSize); i++ {
				var tmpGenerated = &Generated{}
				pb.Next(tmpGenerated, work)
				pb.generatedChan <- tmpGenerated
				sent += 4
			}
			if !reset && sent >= UsedNTimesBufferSize {
				reset = true
				work.Node.GenerateWorkAsync(txCountRI.Next(pb.rng))
			}
		}
	}
}

func (pb *PreviouslyUsedNTimes) Close() {
	close(pb.quitChan)
	pb.waiter.Wait()
}

func (pb *PreviouslyUsedNTimes) GeneratorChan() chan *Generated {
	return pb.generatedChan
}

func (pb *PreviouslyUsedNTimes) ProgressChan() chan utils.Nonce64 {
	return pb.progressChan
}
