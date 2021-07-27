package generators

import (
	"errors"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/rand"
	"sync"
	"time"
)

const (
	DwellingWalkBufferSize         = 16
	DwellingWalkGeneratedCacheSize = 8192
	DwellingTime                   = time.Second * 10
	DwellingExpireTime             = time.Millisecond * 100
	DwellingSteps                  = 4096
	DwellingNtimeReuse             = 64
)

type Dwell struct {
	Nonce      utils.Nonce64
	ExpireChan chan utils.Nonce64
	DeleteChan chan utils.Nonce64
	Expired    bool
	Generated  map[Generated]bool
	Steps      int
	Timer      *time.Timer
	mtx        sync.Mutex
}

type DwellingWalk struct {
	rng            *rand.Rand
	versionSource  *utils.VersionSource
	versions       utils.Versions
	nTime          utils.NTime
	dwell          map[utils.Nonce64]*Dwell
	collisions     []utils.Nonce64
	quitChan       chan struct{}
	versionChan    chan *utils.VersionSource
	workChan       chan int
	generatedChan  chan *Generated
	knownNonceChan chan utils.Nonce64
	waiter         sync.WaitGroup
	progressChan   chan utils.Nonce64
	deleteChan     chan utils.Nonce64
	ntimeUses      int
}

func NewDwell(nonce utils.Nonce64, expireChan chan utils.Nonce64, deleteChan chan utils.Nonce64) *Dwell {
	d := &Dwell{
		Nonce:      nonce,
		ExpireChan: expireChan,
		DeleteChan: deleteChan,
		Generated:  map[Generated]bool{},
	}
	d.Timer = time.AfterFunc(DwellingTime, d.handleExpiration)
	return d
}

func NewDwellingWalk() *DwellingWalk {
	if !ReuseExtraNonce2 {
		panic(errors.New("dwelling requires nonce reuse"))
	}
	var dw = &DwellingWalk{
		rng:            rand.New(rand.NewSource(utils.RandomInt64())),
		dwell:          map[utils.Nonce64]*Dwell{},
		collisions:     []utils.Nonce64{},
		quitChan:       make(chan struct{}),
		versionChan:    make(chan *utils.VersionSource),
		workChan:       make(chan int, DwellingWalkBufferSize),
		generatedChan:  make(chan *Generated, DwellingWalkBufferSize),
		knownNonceChan: make(chan utils.Nonce64, DwellingWalkGeneratedCacheSize),
		progressChan:   make(chan utils.Nonce64, DwellingWalkGeneratedCacheSize),
		deleteChan:     make(chan utils.Nonce64, DwellingWalkGeneratedCacheSize),
	}
	dw.waiter.Add(1)
	go dw.generatorLoop()
	return dw
}

func (d *Dwell) handleExpiration() {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	if d.Expired {
		return
	}
	d.Timer = nil
	d.Expired = true
	d.ExpireChan <- d.Nonce
	d.DeleteChan <- d.Nonce
}

func (d *Dwell) Expire() {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	if d.Timer != nil {
		d.Timer.Reset(DwellingExpireTime)
	}
}

func (dw *DwellingWalk) UpdateVersion(
	versionSource *utils.VersionSource,
) {
	dw.versionChan <- versionSource
}

func (dw *DwellingWalk) UpdateWork() {
	dw.workChan <- 0
}

func (dw *DwellingWalk) ExtraNonceFound(extraNonce utils.Nonce64) {
	dw.knownNonceChan <- extraNonce
}

func (dw *DwellingWalk) Reseed() {
	dw.rng.Seed(utils.RandomInt64())
}

func (dw *DwellingWalk) versionExists(versionPos int, tmpVersion utils.Version) bool {
	for otherPos, otherVersion := range dw.versions {
		if otherPos != versionPos && tmpVersion == otherVersion {
			return true
		}
	}
	return false
}

func (dw *DwellingWalk) Next(generated *Generated) {
	var found bool
	var dwell *Dwell
	for {
		extraNonce2 := utils.Nonce64(utils.RandomUint64())
		if dw.versionSource != nil {
			dw.versionSource.Reset()
		}
		if dwell, found = dw.dwell[extraNonce2]; found && !dwell.Expired {
			dw.NextDwell(dwell, generated)
			return
		} else if !found {
			dwell = NewDwell(extraNonce2, dw.progressChan, dw.deleteChan)
			dw.dwell[extraNonce2] = dwell
			dw.NextDwell(dwell, generated)
			return
		}
	}
}

func (dw *DwellingWalk) NextDwell(dwell *Dwell, generated *Generated) {
	var tmpGenerated Generated
	var found bool
	var versionPos int
	var versions utils.Versions
	var tmpVersions [1]utils.Version

	generated.ExtraNonce2 = dwell.Nonce
	if dw.ntimeUses >= DwellingNtimeReuse {
		dw.nTime = utils.NTime(utils.RandomUint32()) & 0x1ff
		dw.ntimeUses = 0
		if dw.versionSource != nil {
			dw.versionSource.Reset()
		}
	} else {
		dw.ntimeUses += 1
	}
	generated.NTime = dw.nTime
	tmpGenerated.ExtraNonce2 = generated.ExtraNonce2
	tmpGenerated.NTime = generated.NTime

	for versionPos < 4 {
		dw.versionSource.RNGRetrieve(dw.rng, tmpVersions[:])
		tmpGenerated.Version0 = tmpVersions[0]
		if _, found = dwell.Generated[tmpGenerated]; !found {
			dwell.Generated[tmpGenerated] = true
			versions[versionPos] = tmpVersions[0]
			versionPos += 1
		}
	}

	generated.Version1 = versions[1]
	generated.Version0 = versions[0]
	generated.Version2 = versions[2]
	generated.Version3 = versions[3]
	dwell.Steps += 4
}

func (dw *DwellingWalk) generatorLoop() {
	var found bool
	var dwell *Dwell
	var reseedTicker = time.NewTicker(6 * time.Hour)
	var versionSource *utils.VersionSource
	var pending, sent, i, collisionCount, collisionPos, lastSent int
	var knownNonce utils.Nonce64
	for {
		select {
		case <-dw.quitChan:
			dw.waiter.Done()
			return
		case versionSource = <-dw.versionChan:
			dw.versionSource = versionSource
		case <-dw.workChan:
			dw.dwell = map[utils.Nonce64]*Dwell{}
			dw.collisions = []utils.Nonce64{}
			if dw.versionSource != nil {
				dw.versionSource.Reset()
				dw.versionSource.Shuffle(dw.rng)
			}
			dw.nTime = utils.NTime(utils.RandomUint32()) & 0x1ff
			dw.ntimeUses = 0
			collisionPos = 0
			collisionCount = 0
		case <-reseedTicker.C:
			dw.Reseed()
		case knownNonce = <-dw.knownNonceChan:
			if dwell, found = dw.dwell[knownNonce]; found && !dwell.Expired && dwell.Steps < DwellingSteps {
				dwell.Steps = 0
				dw.collisions = append(dw.collisions, knownNonce)
				collisionCount = len(dw.collisions)
			}
			if dw.versionSource != nil {
				dw.versionSource.Reset()
			}
		case knownNonce = <-dw.deleteChan:
			delete(dw.dwell, knownNonce)
		default:
			if versionSource == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			pending = DwellingWalkBufferSize - len(dw.generatedChan)
			if pending == 0 {
				continue
			}
			sent = 0
			lastSent = 0
			if len(dw.collisions) > 0 && len(dw.dwell) > 0 {
				for {
					for _ = range dw.collisions {
						knownNonce = dw.collisions[collisionPos]
						if dwell, found = dw.dwell[knownNonce]; found && !dwell.Expired && dwell.Steps < DwellingSteps {
							if dw.versionSource != nil {
								dw.versionSource.Reset()
							}
							var tmpGenerated = &Generated{}
							dw.NextDwell(dwell, tmpGenerated)
							dw.generatedChan <- tmpGenerated
							sent += 1
							if dwell.Steps >= DwellingSteps {
								dwell.Expire()
							}
						}
						collisionPos += 1
						if collisionPos >= collisionCount {
							collisionPos = 0
						}
						if sent >= pending {
							break
						}
					}
					if sent >= pending || lastSent == sent {
						break
					}
					lastSent = sent
				}
			}
			for i = sent; i < pending; i++ {
				var tmpGenerated = &Generated{}
				dw.Next(tmpGenerated)
				dw.generatedChan <- tmpGenerated
			}
		}
	}
}

func (dw *DwellingWalk) Close() {
	close(dw.quitChan)
	dw.waiter.Wait()
}

func (dw *DwellingWalk) GeneratorChan() chan *Generated {
	return dw.generatedChan
}

func (dw *DwellingWalk) ProgressChan() chan utils.Nonce64 {
	return dw.progressChan
}
