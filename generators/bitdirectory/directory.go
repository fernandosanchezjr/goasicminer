package bitdirectory

func Detail(pos int) (int, int) {
	switch {
	case pos < 8:
		return 0, pos
	case pos < 136:
		var offset = pos - 8
		return (offset / 32) + 1, offset % 32
	case pos < 200:
		return 5, pos - 136
	default:
		return -1, -1
	}
}

func Overview(entry, offset int) int {
	switch {
	case entry == 0:
		return offset
	case entry < 5:
		return ((entry - 1) * 32) + offset + 8
	case entry == 5:
		return 136 + offset
	default:
		return -1
	}
}
