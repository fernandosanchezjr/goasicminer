package protocol

import (
	"encoding/binary"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type Submit struct {
	Difficulty  utils.Difficulty `json:"-"`
	ExtraNonce2 utils.Nonce64    `json:"-"`
	*Method
}

func NewSubmit(
	jobId string,
	extraNonce2 utils.Nonce64,
	ntime utils.NTime,
	nonce utils.Nonce32,
	version utils.Version,
	difficulty utils.Difficulty,
) *Submit {
	var extraNonceB [8]byte
	binary.LittleEndian.PutUint64(extraNonceB[:], uint64(extraNonce2))
	params := []interface{}{
		"",
		jobId,
		fmt.Sprintf("%x", extraNonceB),
		ntime.String(),
		nonce.String(),
	}
	if version != 0 {
		params = append(params, version.String())
	}
	return &Submit{
		difficulty,
		extraNonce2,
		&Method{
			Id:         0,
			MethodName: "mining.submit",
			Params:     params,
		}}
}
