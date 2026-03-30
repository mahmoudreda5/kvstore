package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSetAndGet(t *testing.T) {
	dir := t.TempDir()

	setOutput := captureStdout(t, func () {
		err := run([]string{"kvstore", dir, "set", "name", "mahmoud"})
		if err != nil {
			t.Fatal(err)
		}
	})

	if strings.TrimSpace(setOutput) != `ok set key="name"` {
		t.Fatalf("got %q, want %q", setOutput, `ok set key="name"`)
	}

	getOutput := captureStdout(t, func () {
		err := run([]string {"kvstore", dir, "get", "name"})
		if err != nil {
			t.Fatal(err)
		}
	})

	if strings.TrimSpace(getOutput) != "mahmoud" {
		t.Fatalf("got %q, want %q", getOutput, "mahmoud")
	}
}

func TestRunDeleteThenGetNotFound(t *testing.T) {
	dir := t.TempDir()

	if err := run([]string{"kvstore", dir, "set", "name", "mahmoud"}); err != nil {
		t.Fatal(err)
	}

	deleteOutput := captureStdout(t, func () {
		err := run([]string{"kvstore", dir, "delete", "name"})
		if err != nil {
			t.Fatal(err)
		}
	})

	if strings.TrimSpace(deleteOutput) != `ok delete key="name"` {
		t.Fatalf("got %q, want %q", deleteOutput, `ok delete key="name"`)
	}

	err := run([]string{"kvstore", dir, "get", "name"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != `key "name" not found` {
		t.Fatalf("got %q, want %q", err.Error(), `key "name" not found`)
	}	
}

func TestRunRequiresEnoughArgs(t *testing.T) {
	err := run([]string{"kvstore"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "usage:") {
		t.Fatalf("got %q, want usage error", err.Error())
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout

	tmpFile := filepath.Join(t.TempDir(), "stdout.txt")
	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = file

	fn()

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	os.Stdout = oldStdout

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}
