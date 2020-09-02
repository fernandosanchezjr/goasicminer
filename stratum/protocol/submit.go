package protocol

import (
	"encoding/binary"
	"fmt"
)

type Submit struct {
	*Method
}

func NewSubmit(
	jobId string,
	extraNonce2 uint64,
	ntime uint32,
	nonce uint32,
	version uint32,
) *Submit {
	var extraNonceB [8]byte
	binary.LittleEndian.PutUint64(extraNonceB[:], extraNonce2)
	params := []interface{}{
		"",
		jobId,
		fmt.Sprintf("%x", extraNonceB),
		fmt.Sprintf("%08x", ntime),
		fmt.Sprintf("%08x", nonce),
	}
	if version != 0 {
		params = append(params, fmt.Sprintf("%08x", version))
	}
	return &Submit{&Method{
		Id:         0,
		MethodName: "mining.submit",
		Params:     params,
	}}
}
