package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/epiclabs-io/elastic"
)

type Notify struct {
	JobId          string
	PrevHash       []byte
	CoinBase1      []byte
	CoinBase2      []byte
	MerkleBranches [][]byte
	Version        uint32
	NBits          []byte
	NTime          uint32
	CleanJobs      bool
}

func NewNotify(reply *Reply) (*Notify, error) {
	n := &Notify{}
	if len(reply.Params) != 9 {
		return nil, errors.New("invalid Notify params")
	}
	var prevHash, coinb1, coinb2, version, nbits, ntime string
	var merkleBranch []string
	if err := elastic.Set(&n.JobId, reply.Params[0]); err != nil {
		return nil, err
	}
	if err := elastic.Set(&prevHash, reply.Params[1]); err != nil {
		return nil, err
	}
	if data, err := hex.DecodeString(prevHash); err != nil {
		return nil, err
	} else {
		n.PrevHash = data
	}
	if err := elastic.Set(&coinb1, reply.Params[2]); err != nil {
		return nil, err
	}
	if data, err := hex.DecodeString(coinb1); err != nil {
		return nil, err
	} else {
		n.CoinBase1 = data
	}
	if err := elastic.Set(&coinb2, reply.Params[3]); err != nil {
		return nil, err
	}
	if data, err := hex.DecodeString(coinb2); err != nil {
		return nil, err
	} else {
		n.CoinBase2 = data
	}
	if err := elastic.Set(&merkleBranch, reply.Params[4]); err != nil {
		return nil, err
	}
	n.MerkleBranches = make([][]byte, len(merkleBranch))
	for pos, branch := range merkleBranch {
		if data, err := hex.DecodeString(branch); err != nil {
			return nil, err
		} else {
			n.MerkleBranches[pos] = data
		}
	}
	if err := elastic.Set(&version, reply.Params[5]); err != nil {
		return nil, err
	}
	if data, err := hex.DecodeString(version); err != nil {
		return nil, err
	} else {
		n.Version = binary.BigEndian.Uint32(data)
	}
	if err := elastic.Set(&nbits, reply.Params[6]); err != nil {
		return nil, err
	}
	if data, err := hex.DecodeString(nbits); err != nil {
		return nil, err
	} else {
		n.NBits = data
	}
	if err := elastic.Set(&ntime, reply.Params[7]); err != nil {
		return nil, err
	}
	if data, err := hex.DecodeString(ntime); err != nil {
		return nil, err
	} else {
		n.NTime = binary.BigEndian.Uint32(data)
	}
	if err := elastic.Set(&n.CleanJobs, reply.Params[8]); err != nil {
		return nil, err
	}
	return n, nil
}
