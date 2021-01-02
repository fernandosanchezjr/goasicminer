package uint64

type NibblePositions [16]int

func (np *NibblePositions) shuffler(i, j int) {
	np[i], np[j] = np[j], np[i]
}
