package util

import "encoding/binary"

func Uint64Bytes(val uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(val))
	return buf
}
