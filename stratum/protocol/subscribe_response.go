package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/epiclabs-io/elastic"
	"strings"
)

type SubscribeResponse struct {
	Details        map[string]string
	ExtraNonce1    uint64
	ExtraNonce2Len int
}

func NewSubscribeResponse(reply *Reply) (*SubscribeResponse, error) {
	sr := &SubscribeResponse{
		Details: make(map[string]string),
	}
	if err := reply.HasError(); err != nil {
		return nil, err
	}
	var result []interface{}
	if err := elastic.Set(&result, reply.Result); err != nil {
		return nil, err
	}
	if len(result) != 3 {
		return nil, errors.New("Invalid SubscribeResponse result")
	}
	var rawTuples [][]interface{}
	if err := elastic.Set(&rawTuples, result[0]); err != nil {
		return nil, err
	}
	for _, rawTuple := range rawTuples {
		if len(rawTuple) != 2 {
			return nil, errors.New("Invalid SubscribeResponse result tuple")
		}
		var key, value string
		if err := elastic.Set(&key, rawTuple[0]); err != nil {
			return nil, err
		}
		if err := elastic.Set(&value, rawTuple[1]); err != nil {
			return nil, err
		}
		sr.Details[key] = value
	}
	var hexExtraNonce1 string
	if err := elastic.Set(&hexExtraNonce1, result[1]); err != nil {
		return nil, err
	}
	hexExtraNonce1 = strings.Repeat("0", 16-len(hexExtraNonce1)) + hexExtraNonce1
	if data, err := hex.DecodeString(hexExtraNonce1); err != nil {
		return nil, err
	} else {
		sr.ExtraNonce1 = binary.BigEndian.Uint64(data)
	}
	if err := elastic.Set(&sr.ExtraNonce2Len, result[2]); err != nil {
		return nil, err
	}
	return sr, nil
}
