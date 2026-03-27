//go:build !linux

package main

import (
	"fmt"
	"io"
)

func cmdMount(args []string, w io.Writer) int {
	fmt.Fprintf(w, "mount: FUSE is only supported on Linux.\n")
	fmt.Fprintf(w, "  macOS users: run inside Docker or Multipass.\n")
	return 1
}

func cmdUnmount(args []string, w io.Writer) int {
	fmt.Fprintf(w, "unmount: FUSE is only supported on Linux.\n")
	return 1
}
