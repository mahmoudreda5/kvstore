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

var walMagic = [4]byte{'K', 'V', 'S', 'W'}

const walVersion1 byte = 1

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

func writeWALHeader(w io.Writer) error {
	header := make([]byte, 5)
	copy(header[0:4], walMagic[:])
	header[4] = walVersion1

	_, err := w.Write(header)
	return err
}

func readWALHeader(r io.Reader) (byte, error) {
	var header [5]byte
	_, err := io.ReadFull(r, header[:])
	if err != nil {
		return 0, err
	}

	if string(header[0:4]) != string(walMagic[:]) {
		return 0, errors.New("invalid WAL header")
	}

	return header[4], nil
}
