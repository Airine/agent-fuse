// Package fuse provides the path router and permission model for the agentfs
// virtual filesystem. The router is pure logic with no kernel FUSE dependency,
// so it is fully testable on any platform.
//
// Actual FUSE mount/unmount is in mount_linux.go (Linux only).
package fuse

import (
	"path"
	"strings"
)

// FileKind describes how agentfs treats a virtual path.
type FileKind int

const (
	KindUnknown   FileKind = iota
	KindDirectory          // virtual directories: /, /mem, /skills, /logs
	KindReadOnly           // soul.md, agentfs.yaml, skills/*.md
	KindAppendOnly         // mem/episodic.jsonl
	KindReadWrite          // logs/* files
)

// Route resolves a clean virtual path (e.g. "/soul.md") to its FileKind.
func Route(p string) FileKind {
	p = path.Clean(p)

	switch p {
	case "/", "/mem", "/skills", "/logs":
		return KindDirectory
	case "/soul.md", "/agentfs.yaml":
		return KindReadOnly
	case "/mem/episodic.jsonl":
		return KindAppendOnly
	}

	// skills/*.md — read-only
	if strings.HasPrefix(p, "/skills/") && strings.HasSuffix(p, ".md") {
		return KindReadOnly
	}

	// logs/* — read-write
	if strings.HasPrefix(p, "/logs/") {
		return KindReadWrite
	}

	return KindUnknown
}

// Perm holds the read/write/append capability for a FileKind.
type Perm struct {
	CanRead   bool
	CanWrite  bool
	CanAppend bool
}

// Permissions returns the access capabilities for the given FileKind.
func Permissions(k FileKind) Perm {
	switch k {
	case KindReadOnly, KindDirectory:
		return Perm{CanRead: true}
	case KindAppendOnly:
		return Perm{CanRead: true, CanAppend: true}
	case KindReadWrite:
		return Perm{CanRead: true, CanWrite: true, CanAppend: true}
	default: // KindUnknown
		return Perm{}
	}
}
