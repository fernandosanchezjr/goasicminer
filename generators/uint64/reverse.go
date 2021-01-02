package uint64

import (
	"math/bits"
)

type Reverse struct {
	id uint64
}

func NewReverse() *Reverse {
	return &Reverse{id: NextId()}
}

func (*Reverse) Next(previousState uint64) uint64 {
	return bits.Reverse64(previousState)
}

func (*Reverse) Reseed() {
}

func (r *Reverse) Clone() Generator64 {
	return r
}

func (r *Reverse) Id() uint64 {
	return r.id
}
