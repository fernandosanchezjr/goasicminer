package utils

type MidstateBytes []byte

func (mb MidstateBytes) Reverse() {
	for i, j := 0, len(mb)-1; i < j; i, j = i+1, j-1 {
		mb[i], mb[j] = mb[j], mb[i]
	}
}
