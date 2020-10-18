package generators

type Subtract struct {
}

func (*Subtract) Next(previousState uint64) uint64 {
	return previousState - 1
}
