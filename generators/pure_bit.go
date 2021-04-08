package generators

import (
	"github.com/fernandosanchezjr/goasicminer/generators/bitdirectory"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sort"
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
}

func NewPureBit() *PureBit {
	var pb = &PureBit{
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		overviewRI:     utils.NewRandomIndex(200),
		manipulatedRI:  utils.NewRandomIndex(MaxBitsManipulated - MinBitsManipulated),
		knownNonces:    map[utils.Nonce64]bool{},
		knownGenerated: map[Generated]bool{},
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
	for otherPos, otherVersion := range pb.versions {
		if otherPos != versionPos && tmpVersion == otherVersion {
			return true
		}
	}
	return false
}

func (pb *PureBit) Next(generated *Generated) {
	var tmpGenerated Generated
	var found bool
	var nextBit, entry, offset, pos int
	var tmpVersion utils.Version
	var tmpExtraNonce utils.Nonce64
	var versions utils.Versions
	copy(versions[:], pb.versions[:])
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
				versions[pos] = tmpVersion
			}
		case entry == 5:
			tmpExtraNonce = pb.extraNonce ^ (1 << offset)
			if _, found = pb.knownNonces[tmpExtraNonce]; !found {
				pb.extraNonce = tmpExtraNonce
			}
		}
	}
	generated.ExtraNonce2 = pb.extraNonce
	generated.NTime = pb.nTime
	tmpGenerated = *generated
	copy(pb.versions[:], versions[:])
	sort.Sort(&versions)
	generated.Version0 = versions[0]
	generated.Version1 = versions[1]
	generated.Version2 = versions[2]
	generated.Version3 = versions[3]
	if _, found = pb.knownGenerated[tmpGenerated]; !found {
		pb.knownGenerated[tmpGenerated] = true
	} else {
		pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
		generated.ExtraNonce2 = pb.extraNonce
	}
}

func (pb *PureBit) generatorLoop() {
	var resetBitCountTicker = time.NewTicker(3 * time.Second)
	var reseedTicker = time.NewTicker(2 * time.Minute)
	var versionSource *utils.VersionSource
	var generatedCache = make([]*Generated, GeneratedCacheSize)
	var currentPos, pending, i int
	var knownNonce utils.Nonce64
	var tmpVersion utils.Version
	var sent int
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
			pb.knownNonces = map[utils.Nonce64]bool{}
			pb.knownGenerated = map[Generated]bool{}
			pb.overviewRI.Reset()
			pb.manipulatedRI.Reset()
			pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
			log.WithField("sent", sent).Info("PureBit Generator")
			sent = 0
		case <-reseedTicker.C:
			pb.Reseed()
		case <-resetBitCountTicker.C:
			pb.ResetBitCounts()
		case knownNonce = <-pb.knownNonceChan:
			pb.knownNonces[knownNonce] = true
		default:
			if versionSource == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = BufferSize - len(pb.generatedChan)
			if pending == 0 {
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
			sent += pending
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
