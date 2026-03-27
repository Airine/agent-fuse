//go:build linux

package main

import (
	"fmt"
	"io"

	agentfuse "github.com/aaron/agent-fuse/internal/fuse"
	"github.com/aaron/agent-fuse/internal/manifest"
)

func cmdMount(args []string, w io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintf(w, "Usage: agentfs mount <dir> <mountpoint>\n")
		return 1
	}
	sourceDir, mountpoint := args[0], args[1]

	if _, err := manifest.Parse(sourceDir + "/agentfs.yaml"); err != nil {
		fmt.Fprintf(w, "error: %v\n  hint: run `agentfs init %s` first\n", err, sourceDir)
		return 1
	}

	srv, err := agentfuse.Mount(sourceDir, mountpoint)
	if err != nil {
		fmt.Fprintf(w, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(w, "Mounted %s at %s\n", sourceDir, mountpoint)
	fmt.Fprintf(w, "Press Ctrl-C or run `agentfs unmount %s` to stop\n", mountpoint)
	srv.Wait()
	return 0
}

func cmdUnmount(args []string, w io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintf(w, "Usage: agentfs unmount <mountpoint>\n")
		return 1
	}
	// fusermount -u is the standard unmount on Linux
	fmt.Fprintf(w, "Run: fusermount -u %s\n", args[0])
	return 0
}
