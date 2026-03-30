package main

import (
	"errors"
	"fmt"
	"os"

	"kvstore/internal/store"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 3 {
		return usageError()
	}

	dataDir := args[1]
	command := args[2]

	if command == "help" {
		return usageError()
	}

	s, err := store.Open(dataDir)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	switch command {
	case "set":
		return runSet(s, args)
	case "get":
		return runGet(s, args)
	case "delete":
		return runDelete(s, args)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", command, usageText())
	}
}

func runSet(s *store.Store, args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("usage: kvstore <data-dir> set <key> <value>")
	}

	key := []byte(args[3])
	value := []byte(args[4])

	if err := s.Set(key, value); err != nil {
		return fmt.Errorf("set: %w", err)
	}

	fmt.Printf("ok set key=%q\n", key)
	return nil
}

func runGet(s *store.Store, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("usage: kvstore <data-dir> get <key>")
	}

	key := []byte(args[3])

	value, err := s.Get(key)
	if errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("key %q not found", key)
	}
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}

	fmt.Println(string(value))
	return nil
}

func runDelete(s *store.Store, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("usage: kvstore <data-dir> delete <key>")
	}

	key := []byte(args[3])

	if err := s.Delete(key); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	fmt.Printf("ok delete key=%q\n", key)
	return nil
}

func usageError() error {
	return errors.New(usageText())
}

func usageText() string {
	return "usage:\n" +
		"  kvstore <data-dir> set <key> <value>\n" +
		"  kvstore <data-dir> get <key>\n" +
		"  kvstore <data-dir> delete <key>\n" +
		"  kvstore <data-dir> help"
}
