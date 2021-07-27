package generators

import (
	"github.com/fernandosanchezjr/goasicminer/generators/bitdirectory"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
	"time"
)

const (
	MaxBitsManipulated        = 72
	MinBitsManipulated        = 8
	PureBitBufferSize         = 16
	PureBitNonceReuse         = 64
	PureBitGeneratedCacheSize = 8192
)

type PureBit struct {
	rng             *rand.Rand
	extraNonce      utils.Nonce64
	bitsManipulated int
	overviewRI      *utils.RandomIndex
	manipulatedRI   *utils.RandomIndex
	versionMask     utils.Version
	versions        utils.Versions
	nTime           utils.NTime
	knownNonces     map[utils.Nonce64]bool
	knownGenerated  map[Generated]bool
	quitChan        chan struct{}
	versionChan     chan *utils.VersionSource
	workChan        chan int
	generatedChan   chan *Generated
	knownNonceChan  chan utils.Nonce64
	waiter          sync.WaitGroup
	delayRI         *utils.RandomIndex
	progressChan    chan utils.Nonce64
}

func NewPureBit() *PureBit {
	var pb = &PureBit{
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		manipulatedRI:  utils.NewRandomIndex(MaxBitsManipulated - MinBitsManipulated),
		knownNonces:    map[utils.Nonce64]bool{},
		knownGenerated: map[Generated]bool{},
		quitChan:       make(chan struct{}),
		versionChan:    make(chan *utils.VersionSource),
		workChan:       make(chan int),
		generatedChan:  make(chan *Generated, PureBitBufferSize),
		knownNonceChan: make(chan utils.Nonce64, PureBitGeneratedCacheSize),
		delayRI:        utils.NewRandomIndex(11),
		progressChan:   make(chan utils.Nonce64, PureBitGeneratedCacheSize),
	}
	if !ReuseExtraNonce2 {
		pb.overviewRI = utils.NewRandomIndex(200)
	} else {
		pb.overviewRI = utils.NewRandomIndex(136)
	}
	pb.delayRI.RemovePositions(0)
	pb.bitsManipulated = MinBitsManipulated + pb.manipulatedRI.Next(pb.rng)
	pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
	pb.overviewRI.Shuffle(pb.rng)
	pb.manipulatedRI.Shuffle(pb.rng)
	pb.waiter.Add(1)
	go pb.generatorLoop()
	return pb
}

func (pb *PureBit) UpdateVersion(versionSource *utils.VersionSource) {
	pb.versionChan <- versionSource
}

func (pb *PureBit) UpdateWork() {
	pb.workChan <- 0
}

func (pb *PureBit) ExtraNonceFound(extraNonce utils.Nonce64) {
	pb.knownNonceChan <- extraNonce
}

func (pb *PureBit) Reseed() {
	pb.rng.Seed(utils.RandomInt64())
	pb.overviewRI.Shuffle(pb.rng)
	pb.manipulatedRI.Shuffle(pb.rng)
}

func (pb *PureBit) versionExists(versionPos int, tmpVersion utils.Version) bool {
	for otherPos, otherVersion := range pb.versions {
		if otherPos != versionPos && tmpVersion == otherVersion {
			return true
		}
	}
	return false
}

func (pb *PureBit) Next(generated *Generated) {
	var found bool
	var nextBit, entry, offset, pos int
	var tmpVersion utils.Version
	var tmpExtraNonce utils.Nonce64
	var tmpGenerated Generated
	for {
		for i := 0; i < pb.bitsManipulated; i++ {
			nextBit = pb.overviewRI.Next(pb.rng)
			entry, offset = bitdirectory.Detail(nextBit)
			switch {
			case entry == 0:
				pb.nTime = pb.nTime ^ (1 << offset)
			case entry < 5:
				pos = entry - 1
				tmpVersion = pb.versions[pos] ^ (1 << offset)
				if !pb.versionExists(pos, tmpVersion) {
					pb.versions[pos] = tmpVersion
				} else {
					i -= 1
				}
			case entry == 5:
				tmpExtraNonce = pb.extraNonce ^ (1 << offset)
				if _, found = pb.knownNonces[tmpExtraNonce]; !found {
					pb.extraNonce = tmpExtraNonce
				} else {
					i -= 1
				}
			}
		}
		generated.ExtraNonce2 = pb.extraNonce
		generated.NTime = pb.nTime
		generated.Version0 = pb.versions[0]
		generated.Version1 = pb.versions[1]
		generated.Version2 = pb.versions[2]
		generated.Version3 = pb.versions[3]
		tmpGenerated = *generated
		if _, found = pb.knownGenerated[tmpGenerated]; !found {
			pb.knownGenerated[tmpGenerated] = true
			break
		}
	}
}

func (pb *PureBit) generatorLoop() {
	var resetBitCountTicker = time.NewTicker(1 * time.Second)
	var resetCounts = 0
	var nextReset = pb.delayRI.Next(pb.rng)
	var reseedTicker = time.NewTicker(2 * time.Minute)
	var versionSource *utils.VersionSource
	var generatedCache = make([]*Generated, PureBitGeneratedCacheSize)
	var currentPos, pending, i int
	var knownNonce utils.Nonce64
	var tmpVersion utils.Version
	var reuse int
	for pos := range generatedCache {
		generatedCache[pos] = &Generated{}
	}
	for {
		select {
		case <-pb.quitChan:
			pb.waiter.Done()
			return
		case versionSource = <-pb.versionChan:
			pb.versionMask = versionSource.Version
			var zeroPositions = versionSource.Mask.ZeroPositions()
			for entry := range pb.versions {
				for _, offset := range zeroPositions {
					pb.overviewRI.RemovePositions(bitdirectory.Overview(entry+1, offset))
				}
				for {
					tmpVersion = pb.versionMask | (versionSource.Mask & utils.Version(pb.rng.Uint32()))
					if !pb.versionExists(entry, tmpVersion) {
						pb.versions[entry] = tmpVersion
						break
					}
				}
			}
		case <-pb.workChan:
			pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
			reuse = 0
			pb.knownNonces = map[utils.Nonce64]bool{}
			pb.knownGenerated = map[Generated]bool{}
			pb.delayRI.Reset()
			pb.overviewRI.Reset()
			pb.manipulatedRI.Reset()
			nextReset = pb.delayRI.Next(pb.rng)
		case <-reseedTicker.C:
			pb.Reseed()
		case <-resetBitCountTicker.C:
			if versionSource == nil {
				continue
			}
			resetCounts += 1
			if resetCounts >= nextReset {
				pb.ResetBitCounts()
			}
			if !ReuseExtraNonce2 {
				pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
			}
			for entry := range pb.versions {
				for {
					tmpVersion = pb.versionMask | (versionSource.Mask & utils.Version(pb.rng.Uint32()))
					if !pb.versionExists(entry, tmpVersion) {
						pb.versions[entry] = tmpVersion
						break
					}
				}
			}
		case knownNonce = <-pb.knownNonceChan:
			if !ReuseExtraNonce2 {
				pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
				pb.knownNonces[knownNonce] = true
			}
		default:
			if ReuseExtraNonce2 {
				if reuse >= PureBitNonceReuse {
					extraNonce := pb.extraNonce
					time.AfterFunc(30*time.Millisecond, func() {
						pb.progressChan <- extraNonce
					})
					reuse = 0
					pb.knownNonces[pb.extraNonce] = true
					pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
				}
			}
			if versionSource == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = PureBitBufferSize - len(pb.generatedChan)
			if pending == 0 {
				continue
			}
			for i = 0; i < pending; i++ {
				pb.Next(generatedCache[currentPos])
				pb.generatedChan <- generatedCache[currentPos]
				currentPos += 1
				if currentPos >= PureBitGeneratedCacheSize {
					currentPos = 0
				}
			}
			reuse += pending
		}
	}
}

func (pb *PureBit) ResetBitCounts() {
	if !ReuseExtraNonce2 {
		pb.bitsManipulated = MinBitsManipulated + pb.manipulatedRI.Next(pb.rng)
	} else {
		pb.bitsManipulated = 1
	}
}

func (pb *PureBit) Close() {
	close(pb.quitChan)
	pb.waiter.Wait()
}

func (pb *PureBit) GeneratorChan() chan *Generated {
	return pb.generatedChan
}

func (pb *PureBit) ProgressChan() chan utils.Nonce64 {
	return pb.progressChan
}
