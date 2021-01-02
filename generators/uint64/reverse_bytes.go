package uint64

import (
	"math/bits"
)

type ReverseBytes struct {
	id uint64
}

func NewReverseBytes() *ReverseBytes {
	return &ReverseBytes{id: NextId()}
}

func (*ReverseBytes) Next(previousState uint64) uint64 {
	return bits.ReverseBytes64(previousState)
}

func (*ReverseBytes) Reseed() {
}

func (r *ReverseBytes) Clone() Generator64 {
	return r
}

func (r *ReverseBytes) Id() uint64 {
	return r.id
}
