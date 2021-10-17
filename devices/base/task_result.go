package base

import (
	"github.com/fernandosanchezjr/goasicminer/node"
	"github.com/fernandosanchezjr/goasicminer/utils"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
)

type TaskResult struct {
	Work        *node.Work
	WorkId      uint64
	Version     utils.Version
	VersionPos  int32
	Midstate    int32
	NTime       utils.NTime
	Nonce       utils.Nonce32
	PlainHeader [80]byte
	diffInt     *big.Int
	mtx         sync.Mutex
}

func NewTaskResult() *TaskResult {
	tr := &TaskResult{diffInt: big.NewInt(0)}
	return tr
}

func (tr *TaskResult) doUpdateHeader() {
	// version
	tr.PlainHeader[0] = byte((tr.Version >> 24) & 0xff)
	tr.PlainHeader[1] = byte((tr.Version >> 16) & 0xff)
	tr.PlainHeader[2] = byte((tr.Version >> 8) & 0xff)
	tr.PlainHeader[3] = byte(tr.Version & 0xff)

	// ntime
	tr.PlainHeader[68] = byte((tr.NTime >> 24) & 0xff)
	tr.PlainHeader[69] = byte((tr.NTime >> 16) & 0xff)
	tr.PlainHeader[70] = byte((tr.NTime >> 8) & 0xff)
	tr.PlainHeader[71] = byte(tr.NTime & 0xff)

	// nonce
	tr.PlainHeader[76] = byte((tr.Nonce >> 24) & 0xff)
	tr.PlainHeader[77] = byte((tr.Nonce >> 16) & 0xff)
	tr.PlainHeader[78] = byte((tr.Nonce >> 8) & 0xff)
	tr.PlainHeader[79] = byte(tr.Nonce & 0xff)
}

func (tr *TaskResult) calculateHash() [32]byte {
	tr.doUpdateHeader()
	utils.SwapUint32Bytes(tr.PlainHeader[:])
	return utils.DoubleHash(tr.PlainHeader[:])
}

func (tr *TaskResult) verifyDifficulty(hashBig *big.Int) (reachedMinDifficulty, reachedTargetDifficulty bool) {
	hash := tr.calculateHash()
	if !(hash[31] == 0x0 && hash[30] == 0x0 && hash[29] == 0x0 && hash[28] == 0x0) {
		return false, false
	}
	utils.HashToBig(hash, hashBig)
	if hashBig.Cmp(tr.Work.BigDifficulty) > 0 {
		return false, false
	}
	if hashBig.Cmp(tr.Work.BigTargetDifficulty) <= 0 {
		return true, true
	}
	return true, false
}

func (tr *TaskResult) submit() {
	var work = tr.Work.Clone()
	work.SetNtime(tr.NTime)
	work.SetVersion(tr.Version)
	if submitErr := work.Submit(); submitErr != nil {
		log.WithError(submitErr).Warn("Node submit error")
	} else {
		log.WithFields(log.Fields{
			"jobId":  work.WorkId,
			"height": work.Block.Height(),
		}).Warn("BLOCK MINED")
	}
}

func (tr *TaskResult) Verify(serial string) {
	tr.mtx.Lock()
	defer tr.mtx.Unlock()
	var resultDiff big.Int
	var hashBig big.Int
	var diff utils.Difficulty
	var reachedMinDifficulty, reachedTargetDifficulty = tr.verifyDifficulty(&hashBig)
	if reachedTargetDifficulty {
		tr.submit()
	}
	if reachedMinDifficulty {
		utils.CalculateDifficulty(&hashBig, &resultDiff)
		diff = utils.Difficulty(resultDiff.Int64())
		if diff >= 8192 {
			log.WithFields(log.Fields{
				"serial":       serial,
				"jobId":        tr.WorkId,
				"nTime":        tr.NTime,
				"nonce":        tr.Nonce,
				"version":      tr.Version,
				"difficulty":   diff,
				"transactions": tr.Work.Transactions,
			}).Infoln("Result")
		}
	}
}

func (tr *TaskResult) Lock() {
	tr.mtx.Lock()
}

func (tr *TaskResult) Unlock() {
	tr.mtx.Unlock()
}
