package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
	"time"
)

type PureBit struct {
	rng                *rand.Rand
	extraNonce         utils.Nonce64
	extraNonceReuse    int
	maxExtraNonceReuse int
	extraNonceBitCount int
	extranonceRI       *utils.RandomIndex
	versionMask        utils.Version
	version            [4]utils.Version
	versionBitCount    int
	versionRI          [4]*utils.RandomIndex
	nTime              utils.NTime
	nTimeBitCount      int
	nTimeRI            *utils.RandomIndex
	knownNonces        map[utils.Nonce64]bool
	knownGenerated     map[GeneratedVersion]bool
	quitChan           chan struct{}
	versionChan        chan *utils.VersionSource
	workChan           chan int
	generatedChan      chan *Generated
	knownNonceChan     chan utils.Nonce64
	waiter             sync.WaitGroup
}

func NewPureBit() *PureBit {
	var pb = &PureBit{
		rng:                rand.New(rand.NewSource(utils.RandomInt64())),
		maxExtraNonceReuse: MinExtraNonceReuse,
		extraNonceBitCount: 4,
		extranonceRI:       utils.NewRandomIndex(64),
		versionBitCount:    1,
		nTimeRI:            utils.NewRandomIndex(8),
		nTimeBitCount:      1,
		knownNonces:        map[utils.Nonce64]bool{},
		knownGenerated:     map[GeneratedVersion]bool{},
		quitChan:           make(chan struct{}),
		versionChan:        make(chan *utils.VersionSource),
		workChan:           make(chan int),
		generatedChan:      make(chan *Generated, BufferSize),
		knownNonceChan:     make(chan utils.Nonce64, BufferSize),
	}
	for pos := range pb.versionRI {
		pb.versionRI[pos] = utils.NewRandomIndex(32)
	}
	pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
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
}

func (pb *PureBit) NextExtraNonce() {
	var found bool
	if pb.extraNonceReuse >= pb.maxExtraNonceReuse {
		pb.extraNonceReuse = 0
	}
	if pb.extraNonceReuse == 0 {
		for {
			for i := 0; i < pb.extraNonceBitCount; i++ {
				var nextBit = pb.extranonceRI.Next(pb.rng)
				pb.extraNonce = pb.extraNonce ^ (1 << nextBit)
			}
			if _, found = pb.knownNonces[pb.extraNonce]; !found {
				pb.knownNonces[pb.extraNonce] = true
				break
			}

		}
		pb.nTimeRI.Reset()
		for _, ri := range pb.versionRI {
			ri.Reset()
		}
	}
	pb.extraNonceReuse += 1
}

func (pb *PureBit) NextNTime() {
	for i := 0; i < pb.nTimeBitCount; i++ {
		var nextBit = pb.nTimeRI.Next(pb.rng)
		pb.nTime = pb.nTime ^ (1 << nextBit)
	}
}

func (pb *PureBit) NextVersion() {
	var tmpVersion utils.Version
	var nextBit int
	var added bool
	var ri *utils.RandomIndex
	for versionPos, version := range pb.version {
		added = false
		for added == false {
			ri = pb.versionRI[versionPos]
			for i := 0; i < pb.versionBitCount; i++ {
				nextBit = ri.Next(pb.rng)
				tmpVersion = version ^ (1 << nextBit)
			}
			if !pb.versionExists(versionPos, tmpVersion) {
				pb.version[versionPos] = tmpVersion
				added = true
			}
		}

	}
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
	var found0, found1, found2, found3 bool
	pb.NextExtraNonce()
	generatedHash0.ExtraNonce2, generatedHash1.ExtraNonce2, generatedHash2.ExtraNonce2, generatedHash3.ExtraNonce2 =
		pb.extraNonce, pb.extraNonce, pb.extraNonce, pb.extraNonce
	for {
		pb.NextNTime()
		pb.NextVersion()
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
			break
		}
		pb.extraNonceReuse += 1
	}
}

func (pb *PureBit) generatorLoop() {
	var resetBitCountTicker = time.NewTicker(3 * time.Second)
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
			for _, ri := range pb.versionRI {
				ri.RemovePositions(zeroPositions...)
			}
		case <-pb.workChan:
			pb.extraNonce = utils.Nonce64(pb.rng.Uint64())
			pb.knownNonces = map[utils.Nonce64]bool{}
			pb.knownGenerated = map[GeneratedVersion]bool{}
			pb.extranonceRI.Reset()
		case <-reseedTicker.C:
			pb.Reseed()
		case <-resetBitCountTicker.C:
			pb.ResetBitCounts()
		case knownNonce = <-pb.knownNonceChan:
			if knownNonce == pb.extraNonce {
				pb.extraNonceReuse = MaxExtraNonceReuse
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
	pb.maxExtraNonceReuse = MinExtraNonceReuse + pb.rng.Intn(MaxExtraNonceReuse-MinExtraNonceReuse)
	pb.extraNonceBitCount = 1 + pb.rng.Intn(24)
	pb.versionBitCount = 1 + pb.rng.Intn(4)
	pb.nTimeBitCount = 1 + pb.rng.Intn(3)
}

func (pb *PureBit) Close() {
	close(pb.quitChan)
	pb.waiter.Wait()
}

func (pb *PureBit) GeneratorChan() chan *Generated {
	return pb.generatedChan
}
