package base

import (
	"github.com/fernandosanchezjr/goasicminer/utils"
	"math/big"
)

type TaskResult struct {
	JobId       string
	Version     uint32
	ExtraNonce2 uint64
	NTime       uint32
	Nonce       uint32
	PlainHeader [80]byte
	diffInt     *big.Int
}

func NewTaskResult() *TaskResult {
	tr := &TaskResult{diffInt: big.NewInt(0)}
	return tr
}

func (tr *TaskResult) UpdateHeader() {
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

func (tr *TaskResult) CalculateHash() [32]byte {
	tr.UpdateHeader()
	return utils.DoubleHash(utils.SwapUint32(tr.PlainHeader[:]))
}
