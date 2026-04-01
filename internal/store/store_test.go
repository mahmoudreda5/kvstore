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
	err := os.WriteFile(walPath, encodeRecord(record{
		op: 9,
		key: []byte("name"),
		val: []byte("mahmoud"),
	}), 0o644)
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
	err := os.WriteFile(walPath, truncated, 0o644)
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
