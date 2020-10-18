package generators

type Add struct {
}

func (*Add) Next(previousState uint64) uint64 {
	return previousState + 1
}
