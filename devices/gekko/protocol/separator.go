package protocol

import "bytes"

type MessageSeparator [2]byte

var Separator = MessageSeparator{0x01, 0x60}

func (s MessageSeparator) Search(message []byte) (int, int) {
	var nextPos int
	messageLen := len(message)
	for i := 0; i < messageLen; i++ {
		if message[i] == s[0] {
			nextPos = i + 1
			if nextPos < messageLen && message[nextPos] == s[1] {
				return i, nextPos
			}
		}
	}
	return -1, -1
}

func (s MessageSeparator) Clean(message []byte) []byte {
	var testBuf bytes.Buffer
	data := message[:]
	start, end := Separator.Search(data)
	for start != -1 && end != -1 {
		data = data[end+1:]
		if start, end = Separator.Search(data); start == -1 || end == -1 {
			testBuf.Write(data)
			break
		} else {
			testBuf.Write(data[:start])
		}
	}
	return testBuf.Bytes()
}
