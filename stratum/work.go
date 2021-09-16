package stratum

import (
	"bytes"
	"fmt"
	"github.com/btcsuite/btcutil"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/big"
)

type Work struct {
	ExtraNonce1        uint64
	ExtraNonce2        utils.Nonce64
	ExtraNonce2Len     int
	VersionRolling     bool
	VersionRollingMask utils.Version
	Difficulty         utils.Difficulty
	JobId              string
	PrevHash           [32]byte
	CoinBase1          []byte
	CoinBase2          []byte
	MerkleBranches     [][]byte
	Version            utils.Version
	Nbits              uint32
	Ntime              utils.NTime
	CleanJobs          bool
	Nonce              uint32
	Pool               *Pool
	SubmitChan         chan *protocol.Submit
	TargetDifficulty   utils.Difficulty
	plainHeader        [80]byte
	headerBuf          *bytes.Buffer
	VersionsSource     *utils.VersionSource
	ready              bool
	block              *btcutil.Block
}

type PoolWorkChan chan *Work

func NewWork(
	subscription *protocol.SubscribeResponse,
	configuration *protocol.ConfigureResponse,
	setDifficulty *protocol.SetDifficulty,
	notify *protocol.Notify,
	pool *Pool,
) *Work {
	var result big.Int
	utils.CalculateDifficulty(utils.CompactToBig(notify.NBits), &result)
	w := &Work{
		ExtraNonce1:        subscription.ExtraNonce1,
		ExtraNonce2Len:     subscription.ExtraNonce2Len,
		VersionRolling:     configuration.VersionRolling,
		VersionRollingMask: configuration.VersionRollingMask,
		Difficulty:         setDifficulty.Difficulty,
		JobId:              notify.JobId,
		PrevHash:           notify.PrevHash,
		CoinBase1:          notify.CoinBase1,
		CoinBase2:          notify.CoinBase2,
		MerkleBranches:     notify.MerkleBranches,
		Version:            notify.Version,
		Nbits:              notify.NBits,
		Ntime:              notify.NTime,
		CleanJobs:          notify.CleanJobs,
		Pool:               pool,
		SubmitChan:         pool.SubmitChan,
		TargetDifficulty:   utils.Difficulty(result.Int64()),
		headerBuf:          bytes.NewBuffer(make([]byte, 0, 80)),
	}
	return w
}

func (pw *Work) String() string {
	return fmt.Sprint("Work for block", pw.block.Height())
}

func (pw *Work) PlainHeader() []byte {
	if !pw.ready {
		header := pw.block.MsgBlock().Header
		pw.plainHeader[0] = byte((header.Version >> 24) & 0xff)
		pw.plainHeader[1] = byte((header.Version >> 16) & 0xff)
		pw.plainHeader[2] = byte((header.Version >> 8) & 0xff)
		pw.plainHeader[3] = byte(header.Version & 0xff)
		copy(pw.plainHeader[4:36], header.PrevBlock[:])
		copy(pw.plainHeader[36:68], header.MerkleRoot[:])
		ntime := utils.NTime(header.Timestamp.Unix())
		pw.plainHeader[68] = byte((ntime >> 24) & 0xff)
		pw.plainHeader[69] = byte((ntime >> 16) & 0xff)
		pw.plainHeader[70] = byte((ntime >> 8) & 0xff)
		pw.plainHeader[71] = byte(ntime & 0xff)
		pw.plainHeader[72] = byte((header.Bits >> 24) & 0xff)
		pw.plainHeader[73] = byte((header.Bits >> 16) & 0xff)
		pw.plainHeader[74] = byte((header.Bits >> 8) & 0xff)
		pw.plainHeader[75] = byte(header.Bits & 0xff)
		pw.plainHeader[76] = byte((header.Nonce >> 24) & 0xff)
		pw.plainHeader[77] = byte((header.Nonce >> 16) & 0xff)
		pw.plainHeader[78] = byte((header.Nonce >> 8) & 0xff)
		pw.plainHeader[79] = byte(header.Nonce & 0xff)
		pw.ready = true
	}
	return pw.plainHeader[:]
}

func (pw *Work) Clone() *Work {
	result := *pw
	pw.headerBuf = bytes.NewBuffer(make([]byte, 0, 80))
	return &result
}

func (pw *Work) ExtraNonce2Mask() utils.Nonce64 {
	return 0xffffffffffffffff >> uint64(64-(pw.ExtraNonce2Len*8))
}

func (pw *Work) SetExtraNonce2(extraNonce utils.Nonce64) utils.Nonce64 {
	var nextExtraNonce = extraNonce & pw.ExtraNonce2Mask()
	if pw.ExtraNonce2 != nextExtraNonce {
		pw.ExtraNonce2 = nextExtraNonce
		pw.ready = false
	}
	return pw.ExtraNonce2
}

func (pw *Work) SetNtime(ntime utils.NTime) {
	pw.Ntime = ntime
	pw.plainHeader[68] = byte((pw.Ntime >> 24) & 0xff)
	pw.plainHeader[69] = byte((pw.Ntime >> 16) & 0xff)
	pw.plainHeader[70] = byte((pw.Ntime >> 8) & 0xff)
	pw.plainHeader[71] = byte(pw.Ntime & 0xff)
}
