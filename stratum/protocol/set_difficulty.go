package protocol

import (
	"errors"
	"github.com/dustin/go-humanize"
	"github.com/epiclabs-io/elastic"
	"strconv"
)

const MaxRawDifficulty = 1000

type SetDifficulty struct {
	Difficulty uint64
}

func NewSetDifficulty(reply *Reply) (*SetDifficulty, error) {
	sd := &SetDifficulty{}
	if err := reply.HasError(); err != nil {
		return nil, err
	}
	if len(reply.Params) != 1 {
		return nil, errors.New("invalid SetDifficulty parameters")
	}
	if err := elastic.Set(&sd.Difficulty, reply.Params[0]); err != nil {
		return nil, err
	}
	return sd, nil
}

func (sd *SetDifficulty) String() string {
	if sd.Difficulty < MaxRawDifficulty {
		return strconv.FormatUint(sd.Difficulty, 10)
	} else {
		return humanize.SIWithDigits(float64(sd.Difficulty), 2, "")
	}
}
