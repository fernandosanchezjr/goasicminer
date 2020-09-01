package utils

import (
	"bytes"
	"encoding/binary"
)

func SwapUint32(data []byte) []byte {
	dataLen := len(data)
	readBuf := bytes.NewBuffer(data)
	writeBuf := bytes.NewBuffer(make([]byte, 0, dataLen))
	uint32Count := dataLen / 4
	var tmp uint32
	for i := 0; i < uint32Count; i++ {
		_ = binary.Read(readBuf, binary.LittleEndian, &tmp)
		_ = binary.Write(writeBuf, binary.BigEndian, tmp)
	}
	return writeBuf.Bytes()
}
