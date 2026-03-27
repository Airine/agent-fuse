//go:build linux

package fuse

import (
	"fmt"
	"path/filepath"

	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"

	"github.com/aaron/agent-fuse/internal/manifest"
)

// Mount mounts the agent identity directory at mountpoint.
// Returns an error if agentfs.yaml is missing or invalid.
// Call server.Unmount() to stop.
func Mount(sourceDir, mountpoint string) (*gofuse.Server, error) {
	if _, err := manifest.Parse(filepath.Join(sourceDir, "agentfs.yaml")); err != nil {
		return nil, fmt.Errorf("fuse: %w\n  hint: run `agentfs init %s` first", err, sourceDir)
	}
	fs := &agentFS{root: sourceDir}
	nfs := pathfs.NewPathNodeFs(fs, nil)
	conn := nodefs.NewFileSystemConnector(nfs.Root(), nil)
	server, err := gofuse.NewServer(conn.RawFS(), mountpoint, &gofuse.MountOptions{
		Name:  "agentfs",
		FsName: sourceDir,
	})
	if err != nil {
		return nil, fmt.Errorf("fuse: mount %q → %q: %w", sourceDir, mountpoint, err)
	}
	go server.Serve()
	if err := server.WaitMount(); err != nil {
		return nil, fmt.Errorf("fuse: wait mount: %w", err)
	}
	return server, nil
}
