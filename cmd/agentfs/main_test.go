package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCLI calls run() with the given args and returns the exit code and output.
func runCLI(args ...string) (int, string) {
	var buf strings.Builder
	code := run(args, &buf)
	return code, buf.String()
}

// 1. agentfs init <dir> creates the expected layout
func TestInit_CreatesLayout(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-agent")
	code, out := runCLI("init", dir)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d: %s", code, out)
	}

	paths := []string{
		"agentfs.yaml",
		"soul.md",
		"mem/episodic.jsonl",
		"skills",
		"logs",
	}
	for _, p := range paths {
		if _, err := os.Stat(filepath.Join(dir, p)); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}
}

// 2. agentfs init with no args prints usage and exits non-zero
func TestInit_NoArgs(t *testing.T) {
	code, out := runCLI("init")
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(out, "usage") && !strings.Contains(out, "Usage") {
		t.Errorf("expected usage hint in output, got: %s", out)
	}
}

// 3. agentfs init on non-empty dir prints error and exits non-zero
func TestInit_NonEmptyDir(t *testing.T) {
	dir := t.TempDir()
	// make it non-empty
	os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("x"), 0644)

	code, out := runCLI("init", dir)
	if code == 0 {
		t.Fatal("expected non-zero exit for non-empty dir")
	}
	if !strings.Contains(out, "already exists") && !strings.Contains(out, "not empty") {
		t.Errorf("expected error message, got: %s", out)
	}
}

// 4. unknown command prints usage and exits non-zero
func TestUnknownCommand(t *testing.T) {
	code, out := runCLI("boguscommand")
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown command")
	}
	if !strings.Contains(out, "usage") && !strings.Contains(out, "Usage") && !strings.Contains(out, "unknown") {
		t.Errorf("expected usage/unknown hint in output, got: %s", out)
	}
}

// 5. no args prints usage and exits non-zero
func TestNoArgs(t *testing.T) {
	code, out := runCLI()
	if code == 0 {
		t.Fatal("expected non-zero exit with no args")
	}
	if !strings.Contains(out, "usage") && !strings.Contains(out, "Usage") {
		t.Errorf("expected usage in output, got: %s", out)
	}
}
