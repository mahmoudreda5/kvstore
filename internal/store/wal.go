package store

import (
	"encoding/binary"
	"io"
)

const (
	opSet byte = 1
	opDel byte = 2
)

type record struct {
	op byte
	key []byte
	val []byte
}

func encodeRecord(r record) []byte {
	keyLen := uint32(len(r.key))
	valLen := uint32(len(r.val))

	buf := make([]byte, 1+4+4+keyLen+valLen)
	buf[0] = r.op

	binary.LittleEndian.PutUint32(buf[1:5], keyLen)
	binary.LittleEndian.PutUint32(buf[5:9], valLen)
	copy(buf[9:9+keyLen], r.key)
	copy(buf[9+keyLen:], r.val)

	return buf
}

func decodeRecord(rd io.Reader) (record, error) {
	var header [9]byte
	_, err := io.ReadFull(rd, header[:])
	if err != nil {
		return record{}, err
	}

	op := header[0]
	keyLen := binary.LittleEndian.Uint32(header[1:5])
	valLen := binary.LittleEndian.Uint32(header[5:9])

	key := make([]byte, keyLen)
	if _, err := io.ReadFull(rd, key); err != nil {
		return record{}, err
	}

	val := make([]byte, valLen)
	if _, err := io.ReadFull(rd, val); err != nil {
		return record{}, err
	}

	return record{op: op, key: key, val: val}, nil
}
