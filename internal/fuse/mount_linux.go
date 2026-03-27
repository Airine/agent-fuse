//go:build linux

package fuse

import (
	"fmt"

	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
)

// Mount mounts the agent identity directory at mountpoint.
// Blocks until the server is unmounted. Call server.Unmount() to stop.
func Mount(sourceDir, mountpoint string) (*gofuse.Server, error) {
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
