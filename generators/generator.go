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
	MaxExtraNonceReuse = 1024
	MinExtraNonceReuse = 1024
	MaxVersionReuse    = 32
	MinVersionReuse    = 32
	MaxNTimeReuse      = 32
	MinNTimeReuse      = 32
	BufferSize         = 32
	GeneratedCacheSize = 2048
	Iterations         = 1024
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
	workChan                 chan int
	generatedChan            chan *Generated
	knownNonceChan           chan utils.Nonce64
	waiter                   sync.WaitGroup
}

func NewHeaderFields() *HeaderFields {
	var hf = &HeaderFields{
		extraNonce:            uint642.NewUint64Generator(),
		nTime:                 ntime.NewNtime(),
		version:               version.NewVersion(),
		strategies:            GeneratorStrategies(),
		rng:                   rand.New(rand.NewSource(utils.RandomInt64())),
		quitChan:              make(chan struct{}),
		versionChan:           make(chan *utils.VersionSource),
		workChan:              make(chan int),
		generatedChan:         make(chan *Generated, BufferSize),
		knownNonceChan:        make(chan utils.Nonce64, BufferSize),
		maxStrategyIterations: Iterations,
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
	hf.extraNonce.Shuffle()
	hf.nTime.Shuffle()
	hf.version.Shuffle()
	hf.rng.Shuffle(len(hf.strategies), hf.strategiesShuffler)
}

func (hf *HeaderFields) nextExtraNonce2(strategy Strategy) utils.Nonce64 {
	if strategy == Jump || hf.extraNonceGeneratedCount == 0 || hf.extraNonceGeneratedCount >= hf.maxExtraNonceReuse {
		hf.extraNonceGeneratedCount = 0
		hf.lastExtraNonce = utils.Nonce64(hf.extraNonce.Next())
		hf.version.ResetUsedVersions()
		hf.nTime.ResetUsedNtimes()
	}
	hf.extraNonceGeneratedCount += 1
	return hf.lastExtraNonce
}

func (hf *HeaderFields) nextVersion(strategy Strategy) [4]utils.Version {
	if strategy == Jump || hf.versionGeneratedCount == 0 || hf.versionGeneratedCount >= hf.maxVersionReuse {
		hf.versionGeneratedCount = 0
		hf.lastVersion = hf.version.Next()
	}
	hf.versionGeneratedCount += 1
	return hf.lastVersion
}

func (hf *HeaderFields) nextNTime(strategy Strategy) int {
	if strategy == Jump || hf.nTimeGeneratedCount == 0 || hf.nTimeGeneratedCount >= hf.maxNtimeReuse {
		hf.nTimeGeneratedCount = 0
		hf.lastNTime = hf.nTime.Next()
	}
	hf.nTimeGeneratedCount += 1
	return hf.lastNTime
}

func (hf *HeaderFields) Next(generated *Generated) {
	if hf.strategyIterations >= hf.maxStrategyIterations {
		hf.strategyIterations = 0
		hf.strategyPos += 1
		hf.setMaxReuse()
	}
	if hf.strategyPos >= hf.strategiesCount {
		hf.strategyPos = 0
		hf.Shuffle()
	}
	var strategy = hf.strategies[hf.strategyPos]
	generated.ExtraNonce2 = hf.nextExtraNonce2(Reuse)
	if strategy[0] == Reuse {
		hf.nextVersion(strategy[0])
		generated.NTime = hf.nextNTime(strategy[1])
	} else {
		generated.NTime = hf.nextNTime(strategy[1])
		hf.nextVersion(strategy[0])
	}
	generated.Version0 = hf.lastVersion[0]
	generated.Version1 = hf.lastVersion[1]
	generated.Version2 = hf.lastVersion[2]
	generated.Version3 = hf.lastVersion[3]
	hf.strategyIterations += 1
}

func (hf *HeaderFields) generatorLoop() {
	reseedTicker := time.NewTicker(2 * time.Minute)
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
			hf.version.Reset(versionSource)
		case <-hf.workChan:
			hf.extraNonce.Reset()
		case <-reseedTicker.C:
			hf.Reseed()
		case knownNonce = <-hf.knownNonceChan:
			if knownNonce == hf.lastExtraNonce {
				hf.extraNonceGeneratedCount = MaxExtraNonceReuse
				hf.versionGeneratedCount = MaxVersionReuse
				hf.nTimeGeneratedCount = MaxNTimeReuse
				hf.strategyIterations = 0
				hf.strategyPos += 1
			}
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
