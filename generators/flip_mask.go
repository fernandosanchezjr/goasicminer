package generators

type FlipMask struct {
	mask uint64
}

func NewFlipMask(mask uint64) *FlipMask {
	return &FlipMask{
		mask: mask,
	}
}

func (fm *FlipMask) Next(previousState uint64) uint64 {
	return previousState ^ fm.mask
}

func (fm *FlipMask) Reseed() {

}
