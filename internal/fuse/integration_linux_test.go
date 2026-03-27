//go:build linux

package fuse_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	agentfuse "github.com/aaron/agent-fuse/internal/fuse"
	"github.com/aaron/agent-fuse/internal/scaffold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Full lifecycle: init → mount → read soul.md → append episodic.jsonl → unmount
func TestMount_FullLifecycle(t *testing.T) {
	sourceDir := t.TempDir()
	mountpoint := t.TempDir()

	// 1. Init agent identity
	require.NoError(t, scaffold.Init(sourceDir, "test-agent"))

	// 2. Mount
	srv, err := agentfuse.Mount(sourceDir, mountpoint)
	require.NoError(t, err)
	t.Cleanup(func() { srv.Unmount() })

	// 3. soul.md is readable
	soulPath := filepath.Join(mountpoint, "soul.md")
	data, err := os.ReadFile(soulPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test-agent")

	// 4. soul.md is read-only — write must fail
	err = os.WriteFile(soulPath, []byte("hacked"), 0644)
	assert.Error(t, err, "soul.md should be read-only")

	// 5. episodic.jsonl is writable (append)
	episodicPath := filepath.Join(mountpoint, "mem", "episodic.jsonl")
	f, err := os.OpenFile(episodicPath, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = f.WriteString(`{"ts":"2026-03-27T00:00:00Z","text":"first entry"}` + "\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// 6. episodic.jsonl is readable — entry comes back
	content, err := os.ReadFile(episodicPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "first entry")

	// 7. skills/ is a directory and read-only
	skillsPath := filepath.Join(mountpoint, "skills")
	info, err := os.Stat(skillsPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// 8. logs/ accepts writes
	logPath := filepath.Join(mountpoint, "logs", "session.log")
	err = os.WriteFile(logPath, []byte("agent started\n"), 0644)
	require.NoError(t, err)

	// 9. Unmount cleanly
	require.NoError(t, srv.Unmount())

	// 10. After unmount, source files on disk have the data
	time.Sleep(50 * time.Millisecond) // let VFS flush
	raw, err := os.ReadFile(filepath.Join(sourceDir, "mem", "episodic.jsonl"))
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(raw), "first entry"))
}

// Concurrent appends to episodic.jsonl via FUSE — no interleaved bytes
func TestMount_ConcurrentEpisodicAppend(t *testing.T) {
	sourceDir := t.TempDir()
	mountpoint := t.TempDir()

	require.NoError(t, scaffold.Init(sourceDir, "concurrent-agent"))

	srv, err := agentfuse.Mount(sourceDir, mountpoint)
	require.NoError(t, err)
	t.Cleanup(func() { srv.Unmount() })

	episodicPath := filepath.Join(mountpoint, "mem", "episodic.jsonl")

	done := make(chan error, 20)
	for i := 0; i < 20; i++ {
		go func(n int) {
			f, err := os.OpenFile(episodicPath, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				done <- err
				return
			}
			defer f.Close()
			_, err = f.WriteString(`{"ts":"2026-03-27T00:00:00Z","text":"entry"}` + "\n")
			done <- err
		}(i)
	}

	for i := 0; i < 20; i++ {
		require.NoError(t, <-done)
	}

	content, err := os.ReadFile(episodicPath)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 20)
	for _, line := range lines {
		assert.True(t, strings.HasPrefix(line, "{"), "every line should be valid JSON start: %q", line)
	}
}

// agentfs.yaml missing → Mount returns a clear error
func TestMount_MissingManifest(t *testing.T) {
	sourceDir := t.TempDir() // empty, no agentfs.yaml
	mountpoint := t.TempDir()

	_, err := agentfuse.Mount(sourceDir, mountpoint)
	assert.Error(t, err)
}
