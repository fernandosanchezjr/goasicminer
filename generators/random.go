package generators

import (
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
	"time"
)

const (
	RandomBufferSize         = 64
	RandomGeneratedCacheSize = 8192
	RandomCoinbaseReuse      = 256
	RandomNTimeReuse         = 16
)

type Random struct {
	rng            *rand.Rand
	allVersions    []utils.Version
	versionsRI     *utils.RandomIndex
	nTime          utils.NTime
	nTimeRI        *utils.RandomIndex
	minnTime       utils.NTime
	nTimeReuse     int
	quitChan       chan struct{}
	versionChan    chan *utils.VersionSource
	workChan       chan *node.Work
	generatedChan  chan *Generated
	knownNonceChan chan utils.Nonce64
	waiter         sync.WaitGroup
	progressChan   chan utils.Nonce64
	workId         uint64
}

func NewRandom() *Random {
	var pb = &Random{
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		quitChan:       make(chan struct{}),
		versionChan:    make(chan *utils.VersionSource),
		workChan:       make(chan *node.Work),
		generatedChan:  make(chan *Generated, RandomBufferSize),
		knownNonceChan: make(chan utils.Nonce64, RandomGeneratedCacheSize),
		progressChan:   make(chan utils.Nonce64, RandomGeneratedCacheSize),
		allVersions:    utils.GetUsedVersions(),
	}
	pb.versionsRI = utils.NewRandomIndex(len(pb.allVersions))
	pb.versionsRI.Shuffle(pb.rng)
	pb.waiter.Add(1)
	go pb.generatorLoop()
	return pb
}

func (pb *Random) UpdateVersion(
	versionSource *utils.VersionSource,
) {
	pb.versionChan <- versionSource
}

func (pb *Random) UpdateWork(work *node.Work) {
	pb.workChan <- work
}

func (pb *Random) ExtraNonceFound(extraNonce utils.Nonce64) {
	pb.knownNonceChan <- extraNonce
}

func (pb *Random) Next(generated *Generated, work *node.Work) {
	generated.Work = work.Clone()
	if pb.nTimeReuse >= RandomNTimeReuse {
		pb.nTime = pb.minnTime + utils.NTime(pb.nTimeRI.Next(pb.rng))
		pb.versionsRI.Shuffle(pb.rng)
		pb.nTimeReuse = 0
	}
	generated.NTime = pb.nTime
	//generated.NTime = pb.minnTime + utils.NTime(pb.nTimeRI.Next(pb.rng))

	generated.Work.SetNtime(generated.NTime)

	generated.Version0 = pb.allVersions[pb.versionsRI.Next(pb.rng)]
	generated.Version1 = pb.allVersions[pb.versionsRI.Next(pb.rng)]
	generated.Version2 = pb.allVersions[pb.versionsRI.Next(pb.rng)]
	generated.Version3 = pb.allVersions[pb.versionsRI.Next(pb.rng)]

	pb.nTimeReuse += 4
}

func (pb *Random) generatorLoop() {
	var pending, i int
	var work *node.Work
	var sent int
	var triggered bool
	var txCountRI *utils.RandomIndex
	var txCount int
	for {
		select {
		case <-pb.quitChan:
			pb.waiter.Done()
			return
		case work = <-pb.workChan:
			pb.minnTime = work.MinNtime
			pb.workId = work.WorkId
			if txCountRI == nil || txCount != work.TotalTransactions {
				txCount = work.TotalTransactions
				txCountRI = utils.NewRandomIndex(txCount)
				pb.nTimeRI = utils.NewRandomIndex(int(work.Ntime - pb.minnTime + 300))
				txCountRI.Shuffle(pb.rng)
			}
			pb.nTimeRI.Shuffle(pb.rng)
			pb.versionsRI.Shuffle(pb.rng)
			pb.nTimeReuse = RandomNTimeReuse
			sent = 0
			triggered = false
		default:
			if work == nil || pb.nTimeRI == nil || txCountRI == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = RandomBufferSize - len(pb.generatedChan)
			if pending == 0 {
				continue
			}
			if !triggered && sent >= RandomCoinbaseReuse {
				triggered = true
				work.Node.GenerateWorkAsync(txCountRI.Next(pb.rng))
			}
			for i = 0; i < pending; i++ {
				var tmpGenerated = &Generated{}
				pb.Next(tmpGenerated, work)
				pb.generatedChan <- tmpGenerated
				sent += 4
			}
		}
	}
}

func (pb *Random) Close() {
	close(pb.quitChan)
	pb.waiter.Wait()
}

func (pb *Random) GeneratorChan() chan *Generated {
	return pb.generatedChan
}

func (pb *Random) ProgressChan() chan utils.Nonce64 {
	return pb.progressChan
}
