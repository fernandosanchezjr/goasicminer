package generators

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sync"
	"time"
)

const (
	PureRandomBufferSize         = 16
	PureRandomNonceReuse         = 1024
	PureRandomNtimeReuse         = 32
	PureRandomGeneratedCacheSize = 8192
)

type PureRandom struct {
	rng            *rand.Rand
	extraNonce     utils.Nonce64
	versionSource  *utils.VersionSource
	versions       utils.Versions
	nTime          utils.NTime
	knownNonces    map[utils.Nonce64]bool
	knownGenerated map[Generated]bool
	quitChan       chan struct{}
	versionChan    chan *utils.VersionSource
	workChan       chan int
	generatedChan  chan *Generated
	knownNonceChan chan utils.Nonce64
	waiter         sync.WaitGroup
	progressChan   chan utils.Nonce64
	nonceUses      int
	ntimeUses      int
}

func NewPureRandom() *PureRandom {
	var pb = &PureRandom{
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		knownNonces:    map[utils.Nonce64]bool{},
		knownGenerated: map[Generated]bool{},
		quitChan:       make(chan struct{}),
		versionChan:    make(chan *utils.VersionSource),
		workChan:       make(chan int),
		generatedChan:  make(chan *Generated, PureRandomBufferSize),
		knownNonceChan: make(chan utils.Nonce64, PureRandomGeneratedCacheSize),
		progressChan:   make(chan utils.Nonce64, PureRandomGeneratedCacheSize),
	}
	pb.extraNonce = utils.Nonce64(utils.RandomUint64())
	pb.waiter.Add(1)
	go pb.generatorLoop()
	return pb
}

func (pb *PureRandom) UpdateVersion(
	versionSource *utils.VersionSource,
) {
	pb.versionChan <- versionSource
}

func (pb *PureRandom) UpdateWork() {
	pb.workChan <- 0
}

func (pb *PureRandom) ExtraNonceFound(extraNonce utils.Nonce64) {
	pb.knownNonceChan <- extraNonce
}

func (pb *PureRandom) Reseed() {
	pb.rng.Seed(utils.RandomInt64())
}

func (pb *PureRandom) versionExists(versionPos int, tmpVersion utils.Version) bool {
	for otherPos, otherVersion := range pb.versions {
		if otherPos != versionPos && tmpVersion == otherVersion {
			return true
		}
	}
	return false
}

func (pb *PureRandom) Next(generated *Generated) {
	var tmpGenerated Generated
	var found bool
	var versionPos int
	var versions utils.Versions
	var tmpVersions [1]utils.Version
	if !ReuseExtraNonce2 {
		for {
			pb.extraNonce = utils.Nonce64(utils.RandomUint64())
			if _, found = pb.knownNonces[pb.extraNonce]; !found {
				if pb.versionSource != nil {
					pb.versionSource.Reset()
				}
				break
			}
		}
	}
	generated.ExtraNonce2 = pb.extraNonce
	if !ReuseExtraNonce2 || pb.ntimeUses >= PureRandomNtimeReuse {
		pb.nTime = utils.NTime(utils.RandomUint32()) & 0x1ff
		pb.ntimeUses = 0
		if ReuseExtraNonce2 && pb.versionSource != nil {
			pb.versionSource.Reset()
		}
	} else {
		pb.ntimeUses += 1
	}
	generated.NTime = pb.nTime
	tmpGenerated.ExtraNonce2 = generated.ExtraNonce2
	tmpGenerated.NTime = generated.NTime

	for versionPos < 4 {
		pb.versionSource.RNGRetrieve(pb.rng, tmpVersions[:])
		tmpGenerated.Version0 = tmpVersions[0]
		if _, found = pb.knownGenerated[tmpGenerated]; !found {
			pb.knownGenerated[tmpGenerated] = true
			versions[versionPos] = tmpVersions[0]
			versionPos += 1
		}
	}

	generated.Version1 = versions[1]
	generated.Version0 = versions[0]
	generated.Version2 = versions[2]
	generated.Version3 = versions[3]
	pb.nonceUses += 1
}

func (pb *PureRandom) reset() {
	var extraNonce2 = pb.extraNonce
	time.AfterFunc(30*time.Millisecond, func() {
		pb.progressChan <- extraNonce2
	})
	pb.extraNonce = utils.Nonce64(utils.RandomUint64())
	if pb.versionSource != nil {
		pb.versionSource.Reset()
	}
	if ReuseExtraNonce2 {
		pb.nTime = utils.NTime(utils.RandomUint32()) & 0x1ff
	}
	pb.ntimeUses = 0
}

func (pb *PureRandom) generatorLoop() {
	var reseedTicker = time.NewTicker(6 * time.Hour)
	var versionSource *utils.VersionSource
	var pending, i int
	var knownNonce utils.Nonce64
	var reuse, sent int
	for {
		select {
		case <-pb.quitChan:
			pb.waiter.Done()
			return
		case versionSource = <-pb.versionChan:
			pb.versionSource = versionSource
		case <-pb.workChan:
			pb.knownNonces = map[utils.Nonce64]bool{}
			pb.knownGenerated = map[Generated]bool{}
			pb.extraNonce = utils.Nonce64(utils.RandomUint64())
			if ReuseExtraNonce2 {
				pb.nTime = utils.NTime(utils.RandomUint32()) & 0x1ff
			}
			if pb.versionSource != nil {
				pb.versionSource.Reset()
				pb.versionSource.Shuffle(pb.rng)
			}
			pb.ntimeUses = 0
			log.WithField("sent", sent).Infoln("PureRandom")
			reuse = 0
			sent = 0
		case <-reseedTicker.C:
			pb.Reseed()
		case knownNonce = <-pb.knownNonceChan:
			pb.knownNonces[knownNonce] = true
			if pb.extraNonce == knownNonce {
				pb.reset()
			}
		default:
			if ReuseExtraNonce2 {
				if reuse >= PureRandomNonceReuse {
					reuse = 0
					pb.reset()
				}
			}
			if versionSource == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = PureRandomBufferSize - len(pb.generatedChan)
			if pending == 0 {
				continue
			}
			for i = 0; i < pending; i++ {
				var tmpGenerated = &Generated{}
				pb.Next(tmpGenerated)
				pb.generatedChan <- tmpGenerated
				sent += 4
			}
			reuse += pending * 4
		}
	}
}

func (pb *PureRandom) Close() {
	close(pb.quitChan)
	pb.waiter.Wait()
}

func (pb *PureRandom) GeneratorChan() chan *Generated {
	return pb.generatedChan
}

func (pb *PureRandom) ProgressChan() chan utils.Nonce64 {
	return pb.progressChan
}
