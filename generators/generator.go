package generators

import (
	"github.com/fernandosanchezjr/goasicminer/generators/ntime"
	uint642 "github.com/fernandosanchezjr/goasicminer/generators/uint64"
	"github.com/fernandosanchezjr/goasicminer/generators/version"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
)

const (
	MaxExtraNonceReuse = 32
	MinExtraNonceReuse = 8
	MaxNTimeReuse      = 32
	MinNTimeReuse      = 4
	MaxVersionReuse    = 32
	MinVersionReuse    = 4
)

type HeaderFields struct {
	extraNonce2              *uint642.Uint64
	nTime                    *ntime.NTime
	version                  *version.Version
	rng                      *rand.Rand
	strategies               [][]Strategy
	generated                map[Generated]byte
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
	iterationsPerSecond      int
	maxStrategyIterations    int
}

func NewHeaderFields(iterations int) *HeaderFields {
	var hf = &HeaderFields{
		extraNonce2:         uint642.NewUint64Generator(),
		nTime:               ntime.NewNtime(),
		version:             version.NewVersion(),
		strategies:          GeneratorStrategies(),
		rng:                 rand.New(rand.NewSource(utils.RandomInt64())),
		generated:           map[Generated]byte{},
		iterationsPerSecond: iterations,
	}
	hf.Shuffle()
	hf.setMaxReuse()
	return hf
}

func (hf *HeaderFields) strategiesShuffler(i, j int) {
	hf.strategies[i], hf.strategies[j] = hf.strategies[j], hf.strategies[i]
}

func (hf *HeaderFields) setMaxReuse() {
	hf.maxStrategyIterations = hf.iterationsPerSecond * (1 + hf.rng.Intn(5))
	hf.maxExtraNonceReuse = MaxExtraNonceReuse - hf.rng.Intn(MaxExtraNonceReuse-MinExtraNonceReuse)
	hf.maxNtimeReuse = MinNTimeReuse + hf.rng.Intn(MaxNTimeReuse-MinNTimeReuse)
	hf.maxVersionReuse = MinVersionReuse + hf.rng.Intn(MaxVersionReuse-MinVersionReuse)
}

func (hf *HeaderFields) Reset(nTime utils.NTime, versionSource *utils.VersionSource) {
	hf.nTime.Reset(nTime)
	hf.version.Reset(versionSource)
	hf.generated = map[Generated]byte{}
}

func (hf *HeaderFields) Reseed() {
	hf.rng.Seed(utils.RandomInt64())
	hf.extraNonce2.Reseed()
	hf.nTime.Reseed()
	hf.version.Reseed()
	hf.Shuffle()
}

func (hf *HeaderFields) Shuffle() {
	hf.extraNonce2.Shuffle()
	hf.nTime.Shuffle()
	hf.version.Shuffle()
	hf.rng.Shuffle(len(hf.strategies), hf.strategiesShuffler)
}

func (hf *HeaderFields) nextExtraNonce2(strategy Strategy) utils.Nonce64 {
	if strategy == Jump || hf.extraNonceGeneratedCount == 0 || hf.extraNonceGeneratedCount >= hf.maxExtraNonceReuse {
		hf.extraNonceGeneratedCount = 0
		hf.lastExtraNonce = utils.Nonce64(hf.extraNonce2.Next())
	} else {
		hf.extraNonceGeneratedCount += 1
	}
	return hf.lastExtraNonce
}

func (hf *HeaderFields) nextNTime(strategy Strategy) utils.NTime {
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

func (hf *HeaderFields) Next() (extraNonce2 utils.Nonce64, nTime utils.NTime, versions [4]utils.Version) {
	var generated Generated
	var strategy = hf.strategies[hf.strategyPos]
	for {
		generated.ExtraNonce2 = hf.nextExtraNonce2(strategy[0])
		generated.NTime = hf.nextNTime(strategy[1])
		hf.nextVersion(strategy[2])
		generated.Version0 = hf.lastVersion[0]
		generated.Version1 = hf.lastVersion[1]
		generated.Version2 = hf.lastVersion[2]
		generated.Version3 = hf.lastVersion[3]
		if _, found := hf.generated[generated]; !found {
			hf.generated[generated] = 0
			break
		}
	}
	extraNonce2 = generated.ExtraNonce2
	nTime = generated.NTime
	versions = hf.lastVersion
	hf.strategyIterations += 1
	if hf.strategyIterations >= hf.maxStrategyIterations {
		hf.strategyIterations = 0
		hf.strategyPos += 1
		hf.setMaxReuse()
	}
	if hf.strategyPos >= len(hf.strategies) {
		hf.strategyPos = 0
		hf.Shuffle()
	}
	return
}
