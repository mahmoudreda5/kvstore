package store

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
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

var errInvalidChecksum = errors.New("invalid WAL checksum")

func encodeRecord(r record) []byte {
	keyLen := uint32(len(r.key))
	valLen := uint32(len(r.val))

	payload := make([]byte, 1+4+4+keyLen+valLen)
	payload[0] = r.op

	binary.LittleEndian.PutUint32(payload[1:5], keyLen)
	binary.LittleEndian.PutUint32(payload[5:9], valLen)
	copy(payload[9:9+keyLen], r.key)
	copy(payload[9+keyLen:], r.val)

	checksum := crc32.ChecksumIEEE(payload)

	buf := make([]byte, 4+len(payload))
	binary.LittleEndian.PutUint32(buf[0:4], checksum)
	copy(buf[4:], payload)

	return buf
}

func decodeRecord(rd io.Reader) (record, error) {
	var checksumBuf [4]byte
	_, err := io.ReadFull(rd, checksumBuf[:])
	if err != nil {
		return record{}, err
	}

	expectedChecksum := binary.LittleEndian.Uint32(checksumBuf[:])

	var header [9]byte
	_, err = io.ReadFull(rd, header[:])
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

	payload := make([]byte, 1+4+4+keyLen+valLen)
	copy(payload[0:9], header[:])
	copy(payload[9:9+keyLen], key)
	copy(payload[9+keyLen:], val)

	actualChecksum := crc32.ChecksumIEEE(payload)
	if actualChecksum != expectedChecksum {
		return record{}, errInvalidChecksum
	}

	return record{op: op, key: key, val: val}, nil
}
