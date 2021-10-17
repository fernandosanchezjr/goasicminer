package node

import (
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/big"
	"sync/atomic"
	"time"
)

var workId uint64

type Work struct {
	WorkId              uint64
	Height              int32
	Difficulty          utils.Difficulty
	BigDifficulty       *big.Int
	Version             utils.Version
	Ntime               utils.NTime
	MinNtime            utils.NTime
	Nonce               uint32
	Node                *Node
	TargetDifficulty    utils.Difficulty
	BigTargetDifficulty *big.Int
	plainHeader         [80]byte
	Block               *btcutil.Block
	Transactions        int
	TotalTransactions   int
	ready               bool
}

type WorkChan chan *Work

func NewWork(
	node *Node,
	block *btcutil.Block,
) *Work {
	var header = block.MsgBlock().Header
	var difficulty = utils.Difficulty(1024)
	var tmpDifficulty = big.NewInt(int64(difficulty))
	var bigDifficulty = big.NewInt(0)
	var targetDifficulty = big.NewInt(0)
	var bigTargetDifficulty = utils.CompactToBig(header.Bits)
	utils.CalculateDifficulty(tmpDifficulty, bigDifficulty)
	utils.CalculateDifficulty(bigTargetDifficulty, targetDifficulty)
	w := &Work{
		WorkId:              atomic.AddUint64(&workId, 1),
		Height:              block.Height(),
		Difficulty:          difficulty,
		BigDifficulty:       bigDifficulty,
		Version:             utils.Version(header.Version),
		Ntime:               utils.NTime(node.blockTemplate.CurTime),
		MinNtime:            utils.NTime(node.blockTemplate.MinTime),
		TargetDifficulty:    utils.Difficulty(targetDifficulty.Int64()),
		BigTargetDifficulty: bigTargetDifficulty,
		Node:                node,
		Block:               block,
		Transactions:        len(block.Transactions()),
		TotalTransactions:   len(node.blockTemplate.Transactions),
	}
	return w
}

func (pw *Work) String() string {
	return fmt.Sprint("Work for Block ", pw.Block.Height())
}

func (pw *Work) PlainHeader() []byte {
	if !pw.ready {
		header := pw.Block.MsgBlock().Header
		pw.plainHeader[0] = byte((header.Version >> 24) & 0xff)
		pw.plainHeader[1] = byte((header.Version >> 16) & 0xff)
		pw.plainHeader[2] = byte((header.Version >> 8) & 0xff)
		pw.plainHeader[3] = byte(header.Version & 0xff)
		copy(pw.plainHeader[4:36], header.PrevBlock[:])
		copy(pw.plainHeader[36:68], header.MerkleRoot[:])
		pw.plainHeader[68] = byte((pw.Ntime >> 24) & 0xff)
		pw.plainHeader[69] = byte((pw.Ntime >> 16) & 0xff)
		pw.plainHeader[70] = byte((pw.Ntime >> 8) & 0xff)
		pw.plainHeader[71] = byte(pw.Ntime & 0xff)
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
	var plainHeader [80]byte
	copy(plainHeader[:], pw.plainHeader[:])
	result.plainHeader = plainHeader
	return &result
}

func (pw *Work) SetNtime(ntime utils.NTime) {
	pw.Ntime = ntime
	pw.plainHeader[68] = byte((pw.Ntime >> 24) & 0xff)
	pw.plainHeader[69] = byte((pw.Ntime >> 16) & 0xff)
	pw.plainHeader[70] = byte((pw.Ntime >> 8) & 0xff)
	pw.plainHeader[71] = byte(pw.Ntime & 0xff)
}

func (pw *Work) SetVersion(version utils.Version) {
	pw.Version = version
	pw.plainHeader[0] = byte((pw.Version >> 24) & 0xff)
	pw.plainHeader[1] = byte((pw.Version >> 16) & 0xff)
	pw.plainHeader[2] = byte((pw.Version >> 8) & 0xff)
	pw.plainHeader[3] = byte(pw.Version & 0xff)
}

func (pw *Work) Submit() error {
	var template = pw.Block.MsgBlock().Header
	var msgBlock wire.MsgBlock
	msgBlock.Header = wire.BlockHeader{
		Version:    int32(pw.Version),
		PrevBlock:  template.PrevBlock,
		MerkleRoot: template.MerkleRoot,
		Timestamp:  time.Unix(int64(pw.Ntime), 0),
		Bits:       template.Bits,
	}
	for _, tx := range pw.Block.Transactions() {
		if err := msgBlock.AddTransaction(tx.MsgTx()); err != nil {
			return err
		}
	}
	block := btcutil.NewBlock(&msgBlock)
	block.SetHeight(int32(pw.Block.Height()))
	return pw.Node.Submit(block)
}
