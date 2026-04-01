package main

import (
	"errors"
	"fmt"
	"os"

	"kvstore/internal/store"
)

const (
	exitOK = iota
	exitNotFound
	exitUsage
	exitRuntime
)

type cliError struct {
	code int
	err error
}

func (e *cliError) Error() string {
	return e.err.Error()
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var cliErr *cliError
		if errors.As(err, &cliErr) {
			os.Exit(cliErr.code)
		}
		os.Exit(exitRuntime)
	}
}

func run(args []string) error {
	if len(args) < 3 {
		return usageError()
	}

	dataDir := args[1]
	command := args[2]

	if command == "help" {
		fmt.Println(usageText())
		return nil
	}

	s, err := store.Open(dataDir)
	if err != nil {
		return runtimeErr(fmt.Errorf("open store: %w", err))
	}
	defer s.Close()

	switch command {
	case "set":
		return runSet(s, args)
	case "get":
		return runGet(s, args)
	case "has":
		return runHas(s, args)
	case "delete":
		return runDelete(s, args)
	default:
		return usageErr(fmt.Errorf("unknown command %q\n\n%s", command, usageText()))
	}
}

func runSet(s *store.Store, args []string) error {
	if len(args) != 5 {
		return usageErr(fmt.Errorf("usage: kvstore <data-dir> set <key> <value>"))
	}

	key := []byte(args[3])
	value := []byte(args[4])

	if err := s.Set(key, value); err != nil {
		if errors.Is(err, store.ErrEmptyKey) {
			return usageErr(fmt.Errorf("set: %w", err))
		}
		return runtimeErr(fmt.Errorf("set: %w", err))
	}

	return nil
}

func runGet(s *store.Store, args []string) error {
	if len(args) != 4 {
		return usageErr(fmt.Errorf("usage: kvstore <data-dir> get <key>"))
	}

	key := []byte(args[3])

	value, err := s.Get(key)
	if errors.Is(err, store.ErrNotFound) {
		return notFoundErr(fmt.Errorf("key %q not found", key))
	}
	if err != nil {
		if errors.Is(err, store.ErrEmptyKey) {
			return usageErr(fmt.Errorf("get: %w", err))
		}
		return runtimeErr(fmt.Errorf("get: %w", err))
	}

	fmt.Println(string(value))
	return nil
}

func runDelete(s *store.Store, args []string) error {
	if len(args) != 4 {
		return usageErr(fmt.Errorf("usage: kvstore <data-dir> delete <key>"))
	}

	key := []byte(args[3])

	if err := s.Delete(key); err != nil {
		if errors.Is(err, store.ErrEmptyKey) {
			return usageErr(fmt.Errorf("delete: %w", err))
		}
		return runtimeErr(fmt.Errorf("delete: %w", err))
	}

	return nil
}

func usageError() error {
	return usageErr(errors.New(usageText()))
}

func usageErr(err error) error {
	return &cliError{code: exitUsage, err: err}
}

func notFoundErr(err error) error {
	return &cliError{code: exitNotFound, err: err}
}

func runtimeErr(err error) error {
	return &cliError{code: exitRuntime, err: err}
}

func usageText() string {
	return "usage:\n" +
		"  kvstore <data-dir> set <key> <value>\n" +
		"  kvstore <data-dir> get <key>\n" +
		"  kvstore <data-dir> has <key>\n" +
		"  kvstore <data-dir> delete <key>\n" +
		"  kvstore <data-dir> help"
}

func runHas(s *store.Store, args []string) error {
	if len(args) != 4 {
		return usageErr(fmt.Errorf("usage: kvstore <data-dir> has <key>"))
	}

	key := []byte(args[3])

	_, err := s.Get(key)
	if errors.Is(err, store.ErrNotFound) {
		fmt.Println("false")
		return notFoundErr(errors.New("key not found"))
	}
	if err != nil {
		if errors.Is(err, store.ErrEmptyKey) {
			return usageErr(fmt.Errorf("has: %w", err))
		}
		return runtimeErr(fmt.Errorf("has: %w", err))
	}

	fmt.Println("true")
	return nil
}
