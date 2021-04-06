package generators

import (
	"github.com/fernandosanchezjr/goasicminer/generators/bitdirectory"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
	"time"
)

const MaxBitsManipulated = 96
const MinBitsManipulated = 8

type PureBit struct {
	rng             *rand.Rand
	extraNonce      utils.Nonce64
	bitsManipulated int
	overviewRI      *utils.RandomIndex
	manipulatedRI   *utils.RandomIndex
	versionMask     utils.Version
	version         [4]utils.Version
	nTime           utils.NTime
	knownNonces     map[utils.Nonce64]bool
	knownGenerated  map[GeneratedVersion]bool
	quitChan        chan struct{}
	versionChan     chan *utils.VersionSource
	workChan        chan int
	generatedChan   chan *Generated
	knownNonceChan  chan utils.Nonce64
	waiter          sync.WaitGroup
}

func NewPureBit() *PureBit {
	var pb = &PureBit{
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		overviewRI:     utils.NewRandomIndex(200),
		manipulatedRI:  utils.NewRandomIndex(MaxBitsManipulated - MinBitsManipulated),
		knownNonces:    map[utils.Nonce64]bool{},
		knownGenerated: map[GeneratedVersion]bool{},
		quitChan:       make(chan struct{}),
		versionChan:    make(chan *utils.VersionSource),
		workChan:       make(chan int),
		generatedChan:  make(chan *Generated, BufferSize),
		knownNonceChan: make(chan utils.Nonce64, BufferSize),
	}
	pb.bitsManipulated = pb.manipulatedRI.Next(pb.rng)
	pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
	pb.overviewRI.Shuffle(pb.rng)
	pb.manipulatedRI.Shuffle(pb.rng)
	pb.waiter.Add(1)
	go pb.generatorLoop()
	return pb
}

func (pb *PureBit) IntN(min, max int) int {
	if min == max {
		return min
	}
	return min + pb.rng.Intn(max-min)
}

func (pb *PureBit) UpdateVersion(
	versionSource *utils.VersionSource,
) {
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
	for otherPos, otherVersion := range pb.version {
		if otherPos != versionPos && tmpVersion == otherVersion {
			return true
		}
	}
	return false
}

func (pb *PureBit) Next(generated *Generated) {
	var generatedHash0, generatedHash1, generatedHash2, generatedHash3 GeneratedVersion
	var found, found0, found1, found2, found3 bool
	var nextBit, entry, offset, pos int
	var tmpVersion utils.Version
	var tmpExtraNonce utils.Nonce64
	for {
		for i := 0; i < pb.bitsManipulated; i++ {
			nextBit = pb.overviewRI.Next(pb.rng)
			entry, offset = bitdirectory.Detail(nextBit)
			switch {
			case entry == 0:
				pb.nTime = pb.nTime ^ (1 << offset)
			case entry < 5:
				pos = entry - 1
				tmpVersion = pb.version[pos] ^ (1 << offset)
				if !pb.versionExists(pos, tmpVersion) {
					pb.version[pos] = tmpVersion
				}
			case entry == 5:
				tmpExtraNonce = pb.extraNonce ^ (1 << offset)
				if _, found = pb.knownNonces[tmpExtraNonce]; !found {
					pb.extraNonce = tmpExtraNonce
					pb.knownNonces[tmpExtraNonce] = true
				}
			}
		}
		generatedHash0.ExtraNonce2, generatedHash1.ExtraNonce2, generatedHash2.ExtraNonce2, generatedHash3.ExtraNonce2 =
			pb.extraNonce, pb.extraNonce, pb.extraNonce, pb.extraNonce
		generatedHash0.NTime, generatedHash1.NTime, generatedHash2.NTime, generatedHash3.NTime =
			pb.nTime, pb.nTime, pb.nTime, pb.nTime
		generatedHash0.Version, generatedHash1.Version, generatedHash2.Version, generatedHash3.Version =
			pb.version[0], pb.version[1], pb.version[2], pb.version[3]
		found0 = pb.knownGenerated[generatedHash0]
		found1 = pb.knownGenerated[generatedHash1]
		found2 = pb.knownGenerated[generatedHash2]
		found3 = pb.knownGenerated[generatedHash3]
		if !found0 && !found1 && !found2 && !found3 {
			pb.knownGenerated[generatedHash0] = true
			pb.knownGenerated[generatedHash1] = true
			pb.knownGenerated[generatedHash2] = true
			pb.knownGenerated[generatedHash3] = true
			generated.ExtraNonce2 = pb.extraNonce
			generated.NTime = pb.nTime
			generated.Version0 = pb.versionMask | pb.version[0]
			generated.Version1 = pb.versionMask | pb.version[1]
			generated.Version2 = pb.versionMask | pb.version[2]
			generated.Version3 = pb.versionMask | pb.version[3]
			return
		}
	}
}

func (pb *PureBit) generatorLoop() {
	var resetBitCountTicker = time.NewTicker(1 * time.Second)
	var reseedTicker = time.NewTicker(2 * time.Minute)
	var versionSource *utils.VersionSource
	var generatedCache = make([]*Generated, GeneratedCacheSize)
	var currentPos, pending, i int
	var knownNonce utils.Nonce64
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
			for entry := range pb.version {
				for _, offset := range zeroPositions {
					pb.overviewRI.RemovePositions(bitdirectory.Overview(entry+1, offset))
				}
			}
		case <-pb.workChan:
			pb.knownNonces = map[utils.Nonce64]bool{}
			pb.knownGenerated = map[GeneratedVersion]bool{}
			pb.overviewRI.Reset()
			pb.manipulatedRI.Reset()
		case <-reseedTicker.C:
			pb.Reseed()
		case <-resetBitCountTicker.C:
			pb.ResetBitCounts()
		case knownNonce = <-pb.knownNonceChan:
			if knownNonce == pb.extraNonce {
				pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
				pb.overviewRI.Reset()
			}
		default:
			if versionSource == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = BufferSize - len(pb.generatedChan)
			if pending == 0 {
				time.Sleep(time.Millisecond)
				continue
			}
			for i = 0; i < pending; i++ {
				pb.Next(generatedCache[currentPos])
				pb.generatedChan <- generatedCache[currentPos]
				currentPos += 1
				if currentPos >= GeneratedCacheSize {
					currentPos = 0
				}
			}
		}
	}
}

func (pb *PureBit) ResetBitCounts() {
	pb.bitsManipulated = MinBitsManipulated + pb.manipulatedRI.Next(pb.rng)
}

func (pb *PureBit) Close() {
	close(pb.quitChan)
	pb.waiter.Wait()
}

func (pb *PureBit) GeneratorChan() chan *Generated {
	return pb.generatedChan
}
