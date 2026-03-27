package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aaron/agent-fuse/internal/scaffold"
)

const usage = `Usage: agentfs <command> [args]

Commands:
  init <dir>          Create a new agent identity directory
  mount <dir> <mnt>   Mount an agent identity at <mnt>  [Linux only]
  unmount <mnt>       Unmount an agent identity
`

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

// run executes the CLI with the given args, writing output to w.
// Returns an exit code (0 = success).
func run(args []string, w io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(w, usage)
		return 1
	}

	switch args[0] {
	case "init":
		return cmdInit(args[1:], w)
	case "mount":
		return cmdMount(args[1:], w)
	case "unmount":
		return cmdUnmount(args[1:], w)
	default:
		fmt.Fprintf(w, "unknown command: %q\n\n%s", args[0], usage)
		return 1
	}
}

func cmdInit(args []string, w io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintf(w, "Usage: agentfs init <dir>\n")
		return 1
	}
	dir := args[0]
	name := filepath.Base(dir)
	if err := scaffold.Init(dir, name); err != nil {
		fmt.Fprintf(w, "error: %v\n", err)
		return 1
	}
	fmt.Fprintf(w, "Initialized agent identity at %s\n", dir)
	return 0
}

