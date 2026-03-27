//go:build linux

package fuse

import (
	"context"
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/aaron/agent-fuse/internal/manifest"
	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"
)

// agentNode wraps LoopbackNode and enforces AgentFS path permissions.
type agentNode struct {
	fs.LoopbackNode
}

var _ fs.NodeOpener = (*agentNode)(nil)
var _ fs.NodeGetattrer = (*agentNode)(nil)

// Getattr returns file attributes, stripping write bits on read-only paths.
func (n *agentNode) Getattr(ctx context.Context, fh fs.FileHandle, out *gofuse.AttrOut) syscall.Errno {
	if errno := n.LoopbackNode.Getattr(ctx, fh, out); errno != 0 {
		return errno
	}
	vpath := "/" + n.Path(n.Root())
	perm := Permissions(Route(vpath))
	if !perm.CanWrite && !perm.CanAppend {
		out.Mode &^= 0222 // strip write bits
	}
	return 0
}

// Open checks write permissions before delegating to the loopback.
func (n *agentNode) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	vpath := "/" + n.Path(n.Root())
	perm := Permissions(Route(vpath))

	writing := flags&syscall.O_WRONLY != 0 || flags&syscall.O_RDWR != 0
	if writing && !perm.CanWrite && !perm.CanAppend {
		return nil, 0, syscall.EACCES
	}
	return n.LoopbackNode.Open(ctx, flags)
}

// Mount mounts the agent identity directory at mountpoint.
// Returns an error if agentfs.yaml is missing or invalid.
// Call server.Unmount() to stop.
func Mount(sourceDir, mountpoint string) (*gofuse.Server, error) {
	if _, err := manifest.Parse(filepath.Join(sourceDir, "agentfs.yaml")); err != nil {
		return nil, fmt.Errorf("fuse: %w\n  hint: run `agentfs init %s` first", err, sourceDir)
	}

	rootEmbed, err := fs.NewLoopbackRoot(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("fuse: loopback root: %w", err)
	}
	// NewLoopbackRoot returns a *LoopbackNode as InodeEmbedder.
	// Set NewNode on its RootData so all child lookups use agentNode.
	rootEmbed.(*fs.LoopbackNode).RootData.NewNode = func(
		rootData *fs.LoopbackRoot, parent *fs.Inode, name string, st *syscall.Stat_t,
	) fs.InodeEmbedder {
		return &agentNode{LoopbackNode: fs.LoopbackNode{RootData: rootData}}
	}

	server, err := fs.Mount(mountpoint, rootEmbed, &fs.Options{
		MountOptions: gofuse.MountOptions{
			Name:   "agentfs",
			FsName: sourceDir,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("fuse: mount %q → %q: %w", sourceDir, mountpoint, err)
	}
	return server, nil
}
