package utils

func SwapUint32(data []byte) {
	dataLen := len(data)
	var i, j, k, l int
	for i = 0; i < dataLen; i += 4 {
		j, k, l = i+1, i+2, i+3
		data[i], data[j], data[k], data[l] = data[l], data[k], data[j], data[i]
	}
}
