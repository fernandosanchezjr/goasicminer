package generators

type Zero struct {
}

func (*Zero) Next(uint64) uint64 {
	return 0
}

func (*Zero) Reseed() {
}
