package store

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("key not found")
var ErrEmptyKey = errors.New("key is empty")

type Store struct {
	path string
	data map[string][]byte
	wal  *os.File
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, err
	}

	walPath := filepath.Join(path, "wal.log")
	wal, err := os.OpenFile(walPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	s := &Store{
		path: path,
		data: make(map[string][]byte),
		wal:  wal,
	}

	if err := s.load(); err != nil {
		_ = wal.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) Set(key, value []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}

	rec := record{
		op: opSet,
		key: bytes.Clone(key),
		val: bytes.Clone(value),
	}

	if _, err := s.wal.Write(encodeRecord(rec)); err != nil {
		return err
	}
	if err := s.wal.Sync(); err != nil {
		return err
	}

	s.data[string(key)] = bytes.Clone(value)
	return nil
}

func (s *Store) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, ErrEmptyKey
	}

	value, ok := s.data[string(key)]
	if !ok {
		return nil, ErrNotFound
	}

	return bytes.Clone(value), nil
}

func (s *Store) Delete(key []byte) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}

	rec := record{
		op: opDel,
		key: bytes.Clone(key),
	}

	if _, err := s.wal.Write(encodeRecord(rec)); err != nil {
		return err
	}
	if err := s.wal.Sync(); err != nil {
		return err
	}

	delete(s.data, string(key))
	return nil
}

func (s *Store) Close() error {
	if s.wal == nil {
		return nil
	}

	return s.wal.Close()
}

func (s *Store) load() error {
	if _, err := s.wal.Seek(0, 0); err != nil {
		return err
	}

	for {
		rec, err := decodeRecord(s.wal)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return err
			}
			return err
		}

		switch rec.op {
		case opSet:
			s.data[string(rec.key)] = bytes.Clone(rec.val)
		case opDel:
			delete(s.data, string(rec.key))
		default:
			return fmt.Errorf("unknown WAL op: %d", rec.op)
		}
	}

	_, err := s.wal.Seek(0, io.SeekEnd)
	return err
}
