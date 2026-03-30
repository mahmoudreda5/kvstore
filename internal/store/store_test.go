package store

import (
	"bytes"
	"errors"
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
