package generators

import (
	"github.com/fernandosanchezjr/goasicminer/generators/ntime"
	uint642 "github.com/fernandosanchezjr/goasicminer/generators/uint64"
	"github.com/fernandosanchezjr/goasicminer/generators/version"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
	"time"
)

const (
	MaxExtraNonceReuse = 256
	MinExtraNonceReuse = 64
	MaxVersionReuse    = 16
	MinVersionReuse    = 8
	MaxNTimeReuse      = 16
	MinNTimeReuse      = 8
	BufferSize         = 64
	GeneratedCacheSize = 2048
	Iterations         = 64
)

type HeaderFields struct {
	extraNonce               *uint642.Uint64
	nTime                    *ntime.NTime
	version                  *version.Version
	rng                      *rand.Rand
	strategies               [][]Strategy
	strategiesCount          int
	maxExtraNonceReuse       int
	extraNonceGeneratedCount int
	lastExtraNonce           utils.Nonce64
	maxNtimeReuse            int
	nTimeGeneratedCount      int
	lastNTime                utils.NTime
	maxVersionReuse          int
	versionGeneratedCount    int
	lastVersion              [4]utils.Version
	strategyPos              int
	strategyIterations       int
	started                  bool
	knownGenerated           map[GeneratedVersion]bool
	quitChan                 chan struct{}
	versionChan              chan *utils.VersionSource
	workChan                 chan int
	generatedChan            chan *Generated
	knownNonceChan           chan utils.Nonce64
	waiter                   sync.WaitGroup
}

func NewHeaderFields() *HeaderFields {
	var hf = &HeaderFields{
		extraNonce:     uint642.NewUint64Generator(),
		nTime:          ntime.NewNtime(),
		version:        version.NewVersion(),
		strategies:     GeneratorStrategies(),
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		knownGenerated: map[GeneratedVersion]bool{},
		quitChan:       make(chan struct{}),
		versionChan:    make(chan *utils.VersionSource),
		workChan:       make(chan int),
		generatedChan:  make(chan *Generated, BufferSize),
		knownNonceChan: make(chan utils.Nonce64, BufferSize),
	}
	hf.strategiesCount = len(hf.strategies)
	hf.setMaxReuse()
	hf.Shuffle()
	hf.waiter.Add(1)
	go hf.generatorLoop()
	return hf
}

func (hf *HeaderFields) strategiesShuffler(i, j int) {
	hf.strategies[i], hf.strategies[j] = hf.strategies[j], hf.strategies[i]
}

func (hf *HeaderFields) IntN(min, max int) int {
	if min == max {
		return min
	}
	return min + hf.rng.Intn(max-min)
}

func (hf *HeaderFields) setMaxReuse() {
	hf.maxExtraNonceReuse = hf.IntN(MinExtraNonceReuse, MaxExtraNonceReuse)
	hf.maxNtimeReuse = hf.IntN(MinNTimeReuse, MaxNTimeReuse)
	hf.maxVersionReuse = hf.IntN(MinVersionReuse, MaxVersionReuse)
}

func (hf *HeaderFields) UpdateVersion(
	versionSource *utils.VersionSource,
) {
	hf.versionChan <- versionSource
}

func (hf *HeaderFields) UpdateWork() {
	hf.workChan <- 0
}

func (hf *HeaderFields) ExtraNonceFound(extraNonce utils.Nonce64) {
	hf.knownNonceChan <- extraNonce
}

func (hf *HeaderFields) Reseed() {
	hf.rng.Seed(utils.RandomInt64())
	hf.extraNonce.Reseed()
	hf.nTime.Reseed()
	hf.version.Reseed()
}

func (hf *HeaderFields) Shuffle() {
	hf.rng.Shuffle(len(hf.strategies), hf.strategiesShuffler)
}

func (hf *HeaderFields) nextExtraNonce2(strategy Strategy) utils.Nonce64 {
	if hf.extraNonceGeneratedCount >= hf.maxExtraNonceReuse {
		hf.extraNonceGeneratedCount = 0
	}
	if strategy == Jump || hf.extraNonceGeneratedCount == 0 {
		hf.lastExtraNonce = utils.Nonce64(hf.extraNonce.Next())
		hf.version.Reset()
		hf.nTime.Reset()
		hf.knownGenerated = map[GeneratedVersion]bool{}
		hf.setMaxReuse()
	}
	hf.extraNonceGeneratedCount += 1
	return hf.lastExtraNonce
}

func (hf *HeaderFields) nextVersion(strategy Strategy) [4]utils.Version {
	if hf.versionGeneratedCount >= hf.maxVersionReuse {
		hf.versionGeneratedCount = 0
	}
	if strategy == Jump || hf.versionGeneratedCount == 0 {
		hf.lastVersion = hf.version.Next()
	}
	hf.versionGeneratedCount += 4
	return hf.lastVersion
}

func (hf *HeaderFields) nextNTime(strategy Strategy) utils.NTime {
	if hf.nTimeGeneratedCount >= hf.maxNtimeReuse {
		hf.nTimeGeneratedCount = 0
	}
	if strategy == Jump || hf.nTimeGeneratedCount == 0 {
		hf.lastNTime = hf.nTime.Next()
	}
	hf.nTimeGeneratedCount += 1
	return hf.lastNTime
}

func (hf *HeaderFields) Next(generated *Generated) {
	var generatedHash0, generatedHash1, generatedHash2, generatedHash3 GeneratedVersion
	var found0, found1, found2, found3 bool
	if hf.strategyIterations >= Iterations {
		hf.strategyIterations = 0
		hf.strategyPos += 1
	}
	if hf.strategyPos >= hf.strategiesCount {
		hf.strategyPos = 0
		hf.Shuffle()
	}
	var strategy = hf.strategies[hf.strategyPos]
	generated.ExtraNonce2 = hf.nextExtraNonce2(Reuse)
	for {
		hf.nextVersion(strategy[0])
		generated.NTime = hf.nextNTime(strategy[1])
		generatedHash0.ExtraNonce2, generatedHash1.ExtraNonce2, generatedHash2.ExtraNonce2, generatedHash3.ExtraNonce2 =
			generated.ExtraNonce2, generated.ExtraNonce2, generated.ExtraNonce2, generated.ExtraNonce2
		generatedHash0.NTime, generatedHash1.NTime, generatedHash2.NTime, generatedHash3.NTime =
			generated.NTime, generated.NTime, generated.NTime, generated.NTime
		generatedHash0.Version, generatedHash1.Version, generatedHash2.Version, generatedHash3.Version =
			hf.lastVersion[0], hf.lastVersion[1], hf.lastVersion[2], hf.lastVersion[3]
		found0 = hf.knownGenerated[generatedHash0]
		found1 = hf.knownGenerated[generatedHash1]
		found2 = hf.knownGenerated[generatedHash2]
		found3 = hf.knownGenerated[generatedHash3]
		if !found0 && !found1 && !found2 && !found3 {
			hf.knownGenerated[generatedHash0] = true
			hf.knownGenerated[generatedHash1] = true
			hf.knownGenerated[generatedHash2] = true
			hf.knownGenerated[generatedHash3] = true
			generated.Version0 = hf.lastVersion[0]
			generated.Version1 = hf.lastVersion[1]
			generated.Version2 = hf.lastVersion[2]
			generated.Version3 = hf.lastVersion[3]
			break
		}
		hf.strategyIterations += 1
		hf.extraNonceGeneratedCount += 1
	}
}

func (hf *HeaderFields) generatorLoop() {
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
		case <-hf.quitChan:
			hf.waiter.Done()
			return
		case versionSource = <-hf.versionChan:
			hf.version.Update(versionSource)
		case <-hf.workChan:
			hf.extraNonce.Reset()
		case <-reseedTicker.C:
			hf.Reseed()
		case knownNonce = <-hf.knownNonceChan:
			if knownNonce == hf.lastExtraNonce {
				hf.extraNonceGeneratedCount = MaxExtraNonceReuse
			}
		default:
			if versionSource == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = BufferSize - len(hf.generatedChan)
			if pending == 0 {
				continue
			}
			for i = 0; i < pending; i++ {
				hf.Next(generatedCache[currentPos])
				hf.generatedChan <- generatedCache[currentPos]
				currentPos += 1
				if currentPos >= GeneratedCacheSize {
					currentPos = 0
				}
			}
		}
	}
}

func (hf *HeaderFields) Close() {
	close(hf.quitChan)
	hf.waiter.Wait()
}

func (hf *HeaderFields) GeneratorChan() chan *Generated {
	return hf.generatedChan
}
