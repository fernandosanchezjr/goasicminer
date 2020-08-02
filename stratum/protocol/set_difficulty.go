package protocol

import (
	"errors"
	"github.com/epiclabs-io/elastic"
)

type SetDifficulty struct {
	Difficulty Difficulty
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
	return sd.Difficulty.String()
}
