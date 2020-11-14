package stratum

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	TargetDifficulty   utils.Difficulty
	plainHeader        [80]byte
	headerBuf          *bytes.Buffer
	VersionsSource     *utils.VersionSource
	ready              bool
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
		TargetDifficulty:   utils.Difficulty(result.Int64()),
		headerBuf:          bytes.NewBuffer(make([]byte, 0, 80)),
	}
	return w
}

func (pw *Work) String() string {
	return fmt.Sprint("Work ", pw.JobId, " difficulty ", pw.Difficulty, " from ", pw.Pool)
}

func (pw *Work) Coinbase() []byte {
	extraNonce1 := make([]byte, 8)
	binary.BigEndian.PutUint64(extraNonce1, pw.ExtraNonce1)
	for i := 0; i < 8; i++ {
		if extraNonce1[i] == 0 {
			continue
		} else {
			extraNonce1 = extraNonce1[i:]
			break
		}
	}
	coinbaseLen := len(pw.CoinBase1) + len(extraNonce1) + pw.ExtraNonce2Len + len(pw.CoinBase2)
	coinbaseBytes := make([]byte, 0, coinbaseLen)
	coinbaseBuf := bytes.NewBuffer(coinbaseBytes)
	coinbaseBuf.Write(pw.CoinBase1)
	coinbaseBuf.Write(extraNonce1)
	_ = binary.Write(coinbaseBuf, binary.LittleEndian, pw.ExtraNonce2)
	coinbaseBuf.Write(pw.CoinBase2)
	return coinbaseBuf.Bytes()
}

func (pw *Work) MerkleRoot() []byte {
	var plainText [64]byte
	merkleHash := utils.DoubleHash(pw.Coinbase())
	copy(plainText[0:32], merkleHash[:])
	for _, branch := range pw.MerkleBranches {
		copy(plainText[32:64], branch)
		merkleHash = utils.DoubleHash(plainText[:])
		copy(plainText[0:32], merkleHash[:])
	}
	utils.SwapUint32Bytes(merkleHash[:])
	return merkleHash[:]
}

func (pw *Work) PlainHeader() []byte {
	if !pw.ready {
		pw.plainHeader[0] = byte((pw.Version >> 24) & 0xff)
		pw.plainHeader[1] = byte((pw.Version >> 16) & 0xff)
		pw.plainHeader[2] = byte((pw.Version >> 8) & 0xff)
		pw.plainHeader[3] = byte(pw.Version & 0xff)
		copy(pw.plainHeader[4:36], pw.PrevHash[:])
		copy(pw.plainHeader[36:68], pw.MerkleRoot())
		pw.plainHeader[68] = byte((pw.Ntime >> 24) & 0xff)
		pw.plainHeader[69] = byte((pw.Ntime >> 16) & 0xff)
		pw.plainHeader[70] = byte((pw.Ntime >> 8) & 0xff)
		pw.plainHeader[71] = byte(pw.Ntime & 0xff)
		pw.plainHeader[72] = byte((pw.Nbits >> 24) & 0xff)
		pw.plainHeader[73] = byte((pw.Nbits >> 16) & 0xff)
		pw.plainHeader[74] = byte((pw.Nbits >> 8) & 0xff)
		pw.plainHeader[75] = byte(pw.Nbits & 0xff)
		pw.plainHeader[76] = byte((pw.Nonce >> 24) & 0xff)
		pw.plainHeader[77] = byte((pw.Nonce >> 16) & 0xff)
		pw.plainHeader[78] = byte((pw.Nonce >> 8) & 0xff)
		pw.plainHeader[79] = byte(pw.Nonce & 0xff)
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
	pw.ExtraNonce2 = extraNonce & pw.ExtraNonce2Mask()
	pw.ready = false
	return pw.ExtraNonce2
}

func (pw *Work) SetNtime(ntime utils.NTime) {
	pw.Ntime = ntime
	pw.plainHeader[68] = byte((pw.Ntime >> 24) & 0xff)
	pw.plainHeader[69] = byte((pw.Ntime >> 16) & 0xff)
	pw.plainHeader[70] = byte((pw.Ntime >> 8) & 0xff)
	pw.plainHeader[71] = byte(pw.Ntime & 0xff)
}
