package store

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreSetGet(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	key := []byte("name")
	value := []byte("mahmoud")

	if err := s.Set(key, value); err != nil {
		t.Fatal(err)
	}

	got, err := s.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, value) {
		t.Fatalf("got %q, want %q", got, value)
	}
}

func TestStorePersistsAfterReopen(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	key := []byte("name")
	value := []byte("mahmoud")

	if err := s.Set(key, value); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	got, err := reopened.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, value) {
		t.Fatalf("got %q want %q", got, value)
	}
}

func TestStoreDeletePersistsAfterReopen(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	key := []byte("name")
	value := []byte("mahmoud")

	if err := s.Set(key, value); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(key); err != nil {
		t.Fatal(err)
	}

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	_, err = reopened.Get(key)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got err %v, want %v", err, ErrNotFound)
	}
}

func TestStoreRejectsEmptyKey(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.Set([]byte(""), []byte("value")); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("set: got %v, want %v", err, ErrEmptyKey)
	}

	_, err = s.Get([]byte(""))
	if !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("get: got %v, want %v", err, ErrEmptyKey)
	}

	if err := s.Delete([]byte("")); !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("delete: got %v, want %v", err, ErrEmptyKey)
	}
}

func TestOpenRejectsUnknownWALOp(t *testing.T) {
	dir := t.TempDir()

	walPath := filepath.Join(dir, "wal.log")
	data := make([]byte, 0, 5+32)
	data = append(data, walMagic[:]...)
	data = append(data, walVersion1)
	data = append(data, encodeRecord(record{
		op:  9,
		key: []byte("name"),
		val: []byte("mahmoud"),
	})...)

	err := os.WriteFile(walPath, data, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Open(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unknown WAL op") {
		t.Fatalf("got %q, want unknown WAL op error", err.Error())
	}
}

func TestSetPersistsImmediately(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	key := []byte("name")
	value := []byte("mahmoud")

	if err := s.Set(key, value); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer reopened.Close()
	defer s.Close()

	got, err := reopened.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, value) {
		t.Fatalf("got %q, want %q", got, value)
	}
}

func TestDeletePersistsImmediately(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	key := []byte("name")
	value := []byte("mahmoud")

	if err := s.Set(key, value); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(key); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	defer reopened.Close()
	defer s.Close()

	_, err = reopened.Get(key)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got err %v, want %v", err, ErrNotFound)
	}
}

func TestOpenRejectsTruncatedWALRecord(t *testing.T) {
	dir := t.TempDir()

	full := encodeRecord(record{
		op: opSet,
		key: []byte("name"),
		val: []byte("mahmoud"),
	})

	truncated := full[:len(full)-2]

	walPath := filepath.Join(dir, "wal.log")
	data := make([]byte, 0, 5+len(truncated))
	data = append(data, walMagic[:]...)
	data = append(data, walVersion1)
	data = append(data, truncated...)

	err := os.WriteFile(walPath, data, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Open(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "truncated WAL record") {
		t.Fatalf("got %q, want truncated WAL error", err.Error())
	}
}

func TestOpenRejectsCorruptedWALChecksum(t *testing.T) {
	dir := t.TempDir()

	full := encodeRecord(record{
		op:  opSet,
		key: []byte("name"),
		val: []byte("mahmoud"),
	})

	full[len(full)-1] ^= 0xff

	walPath := filepath.Join(dir, "wal.log")
	data := make([]byte, 0, 5+len(full))
	data = append(data, walMagic[:]...)
	data = append(data, walVersion1)
	data = append(data, full...)

	err := os.WriteFile(walPath, data, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Open(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid WAL checksum") {
		t.Fatalf("got %q, want checksum error", err.Error())
	}
}

func TestStoreHas(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	key := []byte("name")
	value := []byte("mahmoud")

	found, err := s.Has(key)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("expected key to be missing")
	}

	if err := s.Set(key, value); err != nil {
		t.Fatal(err)
	}

	found, err = s.Has(key)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected key to exist")
	}
}

func TestStoreHasRejectsEmptyKey(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	found, err := s.Has([]byte(""))
	if !errors.Is(err, ErrEmptyKey) {
		t.Fatalf("got %v, want %v", err, ErrEmptyKey)
	}
	if found {
		t.Fatal("expected found=false for empty key")
	}
}

func TestOpenWritesWALHeaderForNewStore(t *testing.T) {
	dir := t.TempDir()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	walPath := filepath.Join(dir, "wal.log")
	data, err := os.ReadFile(walPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) < 5 {
		t.Fatalf("wal too short: %d", len(data))
	}
	if string(data[0:4]) != string(walMagic[:]) {
		t.Fatalf("got magic %q, want %q", data[0:4], walMagic[:])
	}
	if data[4] != walVersion1 {
		t.Fatalf("got version %d, want %d", data[4], walVersion1)
	}
}

func TestOpenRejectsInvalidWALHeader(t *testing.T) {
	dir := t.TempDir()

	walPath := filepath.Join(dir, "wal.log")
	err := os.WriteFile(walPath, []byte("BAD!\x02"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Open(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid WAL header") {
		t.Fatalf("got %q, want invalid header error", err.Error())
	}
}

func TestOpenRejectsUnsupportedWALVersion(t *testing.T) {
	dir := t.TempDir()

	walPath := filepath.Join(dir, "wal.log")
	header := []byte{'K', 'V', 'S', 'W', 9}
	err := os.WriteFile(walPath, header, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Open(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported WAL version") {
		t.Fatalf("got %q, want unsupported version error", err.Error())
	}
}
