package generators

type Strategy int

const (
	Reuse Strategy = iota
	Jump
)

func (s Strategy) String() string {
	switch s {
	case Reuse:
		return "Reuse"
	case Jump:
		return "Jump"
	default:
		return "Unknown"
	}
}

func GeneratorStrategies() [][]Strategy {
	var result [][]Strategy
	for i := 1; i < 3; i++ {
		result = append(result, []Strategy{Strategy((i & 2) >> 1), Strategy(i & 1)})
	}
	return result
}
