package inode

import "encoding/binary"

type Inode uint64

func (i Inode) ToBytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func (i Inode) ID() uint64 {
	return uint64(i)
}
