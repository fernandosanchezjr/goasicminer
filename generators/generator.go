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
	MaxExtraNonceReuse = 16
	MinExtraNonceReuse = 4
	MaxNTimeReuse      = 16
	MinNTimeReuse      = 4
	MaxVersionReuse    = 16
	MinVersionReuse    = 4
	BufferSize         = 64
	GeneratedCacheSize = 1024
	Iterations         = 512
)

type HeaderFields struct {
	extraNonce2Source        *uint642.Uint64
	nTime                    *ntime.NTime
	version                  *version.Version
	rng                      *rand.Rand
	strategies               [][]Strategy
	maxExtraNonceReuse       int
	extraNonceGeneratedCount int
	lastExtraNonce           utils.Nonce64
	maxNtimeReuse            int
	nTimeGeneratedCount      int
	lastNTime                int
	maxVersionReuse          int
	versionGeneratedCount    int
	lastVersion              [4]utils.Version
	strategyPos              int
	strategyIterations       int
	maxStrategyIterations    int
	started                  bool
	quitChan                 chan struct{}
	versionChan              chan *utils.VersionSource
	waiter                   sync.WaitGroup
	generatedChan            chan *Generated
}

func NewHeaderFields() *HeaderFields {
	var hf = &HeaderFields{
		extraNonce2Source:     uint642.NewUint64Generator(),
		nTime:                 ntime.NewNtime(),
		version:               version.NewVersion(),
		strategies:            GeneratorStrategies(),
		rng:                   rand.New(rand.NewSource(utils.RandomInt64())),
		quitChan:              make(chan struct{}),
		versionChan:           make(chan *utils.VersionSource),
		generatedChan:         make(chan *Generated, BufferSize),
		maxStrategyIterations: Iterations,
	}
	hf.Shuffle()
	hf.setMaxReuse()
	hf.waiter.Add(1)
	go hf.generatorLoop()
	return hf
}

func (hf *HeaderFields) strategiesShuffler(i, j int) {
	hf.strategies[i], hf.strategies[j] = hf.strategies[j], hf.strategies[i]
}

func (hf *HeaderFields) setMaxReuse() {
	hf.maxExtraNonceReuse = MinExtraNonceReuse + hf.rng.Intn(MaxExtraNonceReuse-MinExtraNonceReuse)
	hf.maxNtimeReuse = MinNTimeReuse + hf.rng.Intn(MaxNTimeReuse-MinNTimeReuse)
	hf.maxVersionReuse = MinVersionReuse + hf.rng.Intn(MaxVersionReuse-MinVersionReuse)
}

func (hf *HeaderFields) Reset(
	versionSource *utils.VersionSource,
) {
	hf.versionChan <- versionSource
}

func (hf *HeaderFields) Reseed() {
	hf.rng.Seed(utils.RandomInt64())
	if hf.extraNonce2Source != nil {
		hf.extraNonce2Source.Reseed()
	}
	hf.nTime.Reseed()
	hf.version.Reseed()
}

func (hf *HeaderFields) Shuffle() {
	hf.extraNonce2Source.Shuffle()
	hf.nTime.Shuffle()
	hf.version.Shuffle()
}

func (hf *HeaderFields) nextExtraNonce2(strategy Strategy) utils.Nonce64 {
	if strategy == Jump || hf.extraNonceGeneratedCount == 0 || hf.extraNonceGeneratedCount >= hf.maxExtraNonceReuse {
		hf.extraNonceGeneratedCount = 0
		hf.lastExtraNonce = utils.Nonce64(hf.extraNonce2Source.Next())
	} else {
		hf.extraNonceGeneratedCount += 1
	}
	return hf.lastExtraNonce
}

func (hf *HeaderFields) nextNTime(strategy Strategy) int {
	if strategy == Jump || hf.nTimeGeneratedCount == 0 || hf.nTimeGeneratedCount >= hf.maxNtimeReuse {
		hf.nTimeGeneratedCount = 0
		hf.lastNTime = hf.nTime.Next()
	} else {
		hf.nTimeGeneratedCount += 1
	}
	return hf.lastNTime
}

func (hf *HeaderFields) nextVersion(strategy Strategy) [4]utils.Version {
	if strategy == Jump || hf.versionGeneratedCount == 0 || hf.versionGeneratedCount >= hf.maxVersionReuse {
		hf.versionGeneratedCount = 0
		hf.lastVersion = hf.version.Next()
	} else {
		hf.versionGeneratedCount += 1
	}
	return hf.lastVersion
}

func (hf *HeaderFields) Next(generated *Generated) {
	if hf.strategyIterations >= hf.maxStrategyIterations {
		hf.setMaxReuse()
		hf.Shuffle()
		hf.strategyIterations = 0
		hf.strategyPos += 1
	}
	if hf.strategyPos >= len(hf.strategies) {
		hf.strategyPos = 0
		hf.rng.Shuffle(len(hf.strategies), hf.strategiesShuffler)
	}
	var strategy = hf.strategies[hf.strategyPos]
	generated.ExtraNonce2 = hf.nextExtraNonce2(strategy[0])
	generated.NTime = hf.nextNTime(strategy[1])
	hf.nextVersion(strategy[2])
	generated.Version0 = hf.lastVersion[0]
	generated.Version1 = hf.lastVersion[1]
	generated.Version2 = hf.lastVersion[2]
	generated.Version3 = hf.lastVersion[3]
	hf.strategyIterations += 1
}

func (hf *HeaderFields) generatorLoop() {
	reseedTicker := time.NewTicker(1 * time.Minute)
	var versionSource *utils.VersionSource
	var generatedCache = make([]*Generated, GeneratedCacheSize)
	var currentPos, pending, i int
	for pos := range generatedCache {
		generatedCache[pos] = &Generated{}
	}
	for {
		select {
		case <-hf.quitChan:
			hf.waiter.Done()
			return
		case versionSource = <-hf.versionChan:
			hf.extraNonce2Source.Reset()
			hf.version.Reset(versionSource)
		case <-reseedTicker.C:
			hf.Reseed()
		default:
			if versionSource == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = BufferSize - len(hf.generatedChan)
			if pending == 0 {
				time.Sleep(time.Millisecond)
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
