//go:build linux

package fuse

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
)

// agentFS implements pathfs.FileSystem, routing each path through the router
// and enforcing read/write permissions.
type agentFS struct {
	pathfs.FileSystem
	root string
}

func (fs *agentFS) realPath(name string) string {
	return filepath.Join(fs.root, name)
}

func (fs *agentFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	vpath := "/" + name
	kind := Route(vpath)
	if kind == KindUnknown {
		return nil, fuse.ENOENT
	}

	if kind == KindDirectory {
		return &fuse.Attr{Mode: fuse.S_IFDIR | 0555}, fuse.OK
	}

	fi, err := os.Stat(fs.realPath(name))
	if err != nil {
		return nil, fuse.ENOENT
	}

	perm := Permissions(kind)
	mode := uint32(fuse.S_IFREG)
	if perm.CanWrite || perm.CanAppend {
		mode |= 0644
	} else {
		mode |= 0444
	}

	return &fuse.Attr{
		Mode:  mode,
		Size:  uint64(fi.Size()),
		Mtime: uint64(fi.ModTime().Unix()),
	}, fuse.OK
}

func (fs *agentFS) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	vpath := "/" + name
	kind := Route(vpath)
	if kind != KindDirectory {
		return nil, fuse.ENOTDIR
	}

	entries, err := os.ReadDir(fs.realPath(name))
	if err != nil {
		return nil, fuse.OK // empty virtual dir is fine
	}

	var out []fuse.DirEntry
	for _, e := range entries {
		childPath := vpath + "/" + e.Name()
		if e.IsDir() {
			childPath = "/" + e.Name()
		}
		childKind := Route(childPath)
		if childKind == KindUnknown {
			continue // hide unknown entries
		}
		mode := uint32(fuse.S_IFREG)
		if e.IsDir() {
			mode = fuse.S_IFDIR
		}
		out = append(out, fuse.DirEntry{Name: e.Name(), Mode: mode})
	}
	return out, fuse.OK
}

func (fs *agentFS) Open(name string, flags uint32, context *fuse.Context) (pathfs.File, fuse.Status) {
	vpath := "/" + name
	kind := Route(vpath)
	perm := Permissions(kind)

	writing := flags&syscall.O_WRONLY != 0 || flags&syscall.O_RDWR != 0
	appending := flags&syscall.O_APPEND != 0

	if writing && !perm.CanWrite && !perm.CanAppend {
		return nil, fuse.EACCES
	}
	if appending && !perm.CanAppend {
		return nil, fuse.EACCES
	}

	f, err := os.OpenFile(fs.realPath(name), int(flags), 0644)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	return pathfs.NewLoopbackFile(f), fuse.OK
}

func (fs *agentFS) Create(name string, flags uint32, mode uint32, context *fuse.Context) (pathfs.File, fuse.Status) {
	vpath := "/" + name
	kind := Route(vpath)
	perm := Permissions(kind)
	if !perm.CanWrite && !perm.CanAppend {
		return nil, fuse.EACCES
	}
	f, err := os.OpenFile(fs.realPath(name), os.O_CREATE|int(flags), os.FileMode(mode))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	return pathfs.NewLoopbackFile(f), fuse.OK
}

func (fs *agentFS) Unlink(name string, context *fuse.Context) fuse.Status {
	return fuse.EACCES // no deletions via FUSE
}

func (fs *agentFS) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	return fuse.EACCES
}

func (fs *agentFS) Rmdir(name string, context *fuse.Context) fuse.Status {
	return fuse.EACCES
}

func (fs *agentFS) Rename(oldName, newName string, context *fuse.Context) fuse.Status {
	return fuse.EACCES
}

func (fs *agentFS) String() string { return "agentfs" }
