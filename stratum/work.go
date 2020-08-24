package stratum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/stratum/protocol"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/bits"
)

type Work struct {
	ExtraNonce1        uint64
	ExtraNonce2        uint64
	ExtraNonce2Len     int
	VersionRolling     bool
	VersionRollingMask uint32
	Difficulty         protocol.Difficulty
	JobId              string
	PrevHash           []byte
	CoinBase1          []byte
	CoinBase2          []byte
	MerkleBranches     [][]byte
	Version            uint32
	Nbits              []byte
	Ntime              uint32
	CleanJobs          bool
	Nonce              uint32
	Pool               *Pool
	plainHeader        []byte
	versions           []uint32
}

type PoolWorkChan chan *Work

func NewWork(
	subscription *protocol.SubscribeResponse,
	configuration *protocol.ConfigureResponse,
	setDifficulty *protocol.SetDifficulty,
	notify *protocol.Notify,
	pool *Pool,
) *Work {
	return &Work{
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
	}
}

func (pw *Work) String() string {
	return fmt.Sprint("Work ", pw.JobId, " difficulty ", pw.Difficulty, " from ", pw.Pool)
}

func (pw *Work) Coinbase() []byte {
	coinbaseLen := len(pw.CoinBase1) + 8 + pw.ExtraNonce2Len + len(pw.CoinBase2)
	coinbaseBytes := make([]byte, 0, coinbaseLen)
	coinbaseBuf := bytes.NewBuffer(coinbaseBytes)
	coinbaseBuf.Write(pw.CoinBase1)
	_ = binary.Write(coinbaseBuf, binary.BigEndian, pw.ExtraNonce1)
	_ = binary.Write(coinbaseBuf, binary.BigEndian, pw.ExtraNonce2)
	coinbaseBuf.Write(pw.CoinBase2)
	return coinbaseBuf.Bytes()
}

func (pw *Work) MerkleRoot() []byte {
	coinbase := utils.DoubleHash(pw.Coinbase())
	plainText := make([]byte, 64)
	merkle_root := coinbase
	for _, branch := range pw.MerkleBranches {
		copy(plainText, merkle_root[:])
		copy(plainText[32:], branch)
		merkle_root = utils.DoubleHash(plainText)
	}
	return merkle_root[:]
}

func (pw *Work) PlainHeader() []byte {
	if len(pw.plainHeader) == 0 {
		headerBuf := bytes.NewBuffer(make([]byte, 0, 80))
		_ = binary.Write(headerBuf, binary.BigEndian, pw.Version)
		headerBuf.Write(pw.PrevHash)
		headerBuf.Write(pw.MerkleRoot())
		_ = binary.Write(headerBuf, binary.BigEndian, pw.Ntime)
		headerBuf.Write(pw.Nbits)
		_ = binary.Write(headerBuf, binary.BigEndian, pw.Nonce)
		pw.plainHeader = headerBuf.Bytes()
	}
	return pw.plainHeader
}

func (pw *Work) Versions(maxCount int) []uint32 {
	if len(pw.versions) == 0 {
		// Inspired by docs from https://github.com/slushpool/stratumprotocol/blob/master/stratum-extensions.mediawiki
		tmpMask := pw.VersionRollingMask
		maxOnes := bits.OnesCount32(tmpMask)
		if maxCount > maxOnes {
			maxCount = maxOnes
		}
		pw.versions = make([]uint32, 0, maxCount)
		pw.versions = append(pw.versions, pw.Version)
		for i := 0; i < 32; i++ {
			if (tmpMask & 1) == 1 {
				pw.versions = append(pw.versions, pw.Version|1<<i)
				if len(pw.versions) == maxCount {
					break
				}
			}
			tmpMask = tmpMask >> 1
		}
	}
	return pw.versions
}

func (pw *Work) Clone() *Work {
	result := *pw
	return &result
}

func (pw *Work) Reset() {
	pw.plainHeader = nil
	pw.versions = nil
}

func (pw *Work) ExtraNonce2Mask() uint64 {
	return 0xffffffffffffffff >> uint64(64-(pw.ExtraNonce2Len*8))
}

func (pw *Work) SetExtraNonce2(extraNonce uint64) {
	pw.ExtraNonce2 = extraNonce & pw.ExtraNonce2Mask()
	pw.Reset()
}
