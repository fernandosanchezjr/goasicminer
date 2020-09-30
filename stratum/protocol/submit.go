package protocol

import (
	"encoding/binary"
	"fmt"
	"github.com/fernandosanchezjr/goasicminer/utils"
)

type Submit struct {
	*Method
}

func NewSubmit(
	jobId string,
	extraNonce2 utils.Nonce64,
	ntime utils.NTime,
	nonce utils.Nonce32,
	version utils.Version,
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
	return &Submit{&Method{
		Id:         0,
		MethodName: "mining.submit",
		Params:     params,
	}}
}
