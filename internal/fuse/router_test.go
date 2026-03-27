package fuse_test

import (
	"testing"

	agentfuse "github.com/aaron/agent-fuse/internal/fuse"
	"github.com/stretchr/testify/assert"
)

// Route maps virtual paths to access kinds.
// This is pure logic — no kernel FUSE needed.

func TestRoute_ReadOnlyPaths(t *testing.T) {
	readOnly := []string{
		"/soul.md",
		"/agentfs.yaml",
		"/skills/web-search.md",
		"/skills/code-review.md",
	}
	for _, p := range readOnly {
		assert.Equal(t, agentfuse.KindReadOnly, agentfuse.Route(p), "path: %s", p)
	}
}

func TestRoute_AppendOnlyPaths(t *testing.T) {
	appendOnly := []string{
		"/mem/episodic.jsonl",
	}
	for _, p := range appendOnly {
		assert.Equal(t, agentfuse.KindAppendOnly, agentfuse.Route(p), "path: %s", p)
	}
}

func TestRoute_ReadWritePaths(t *testing.T) {
	readWrite := []string{
		"/logs/session.log",
		"/logs/2026-03-27.log",
	}
	for _, p := range readWrite {
		assert.Equal(t, agentfuse.KindReadWrite, agentfuse.Route(p), "path: %s", p)
	}
}

func TestRoute_DirectoryPaths(t *testing.T) {
	dirs := []string{
		"/",
		"/mem",
		"/skills",
		"/logs",
	}
	for _, p := range dirs {
		assert.Equal(t, agentfuse.KindDirectory, agentfuse.Route(p), "path: %s", p)
	}
}

func TestRoute_UnknownPaths(t *testing.T) {
	unknown := []string{
		"/unknown.txt",
		"/mem/other.txt",
		"/etc/passwd",
	}
	for _, p := range unknown {
		assert.Equal(t, agentfuse.KindUnknown, agentfuse.Route(p), "path: %s", p)
	}
}

// Permissions derives whether read/write is allowed for a given kind.
func TestPermissions(t *testing.T) {
	tests := []struct {
		kind      agentfuse.FileKind
		canRead   bool
		canWrite  bool
		canAppend bool
	}{
		{agentfuse.KindReadOnly, true, false, false},
		{agentfuse.KindAppendOnly, true, false, true},
		{agentfuse.KindReadWrite, true, true, true},
		{agentfuse.KindDirectory, true, false, false},
		{agentfuse.KindUnknown, false, false, false},
	}
	for _, tt := range tests {
		p := agentfuse.Permissions(tt.kind)
		assert.Equal(t, tt.canRead, p.CanRead, "kind %v CanRead", tt.kind)
		assert.Equal(t, tt.canWrite, p.CanWrite, "kind %v CanWrite", tt.kind)
		assert.Equal(t, tt.canAppend, p.CanAppend, "kind %v CanAppend", tt.kind)
	}
}
