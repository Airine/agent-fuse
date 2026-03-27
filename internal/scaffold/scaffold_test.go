package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aaron/agent-fuse/internal/scaffold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test 1: Init a new directory → creates all 6 items.
func TestInit_CreatesAllItems(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myagent")

	err := scaffold.Init(dir, "myagent")
	require.NoError(t, err)

	entries := []string{
		"agentfs.yaml",
		"soul.md",
		"mem",
		filepath.Join("mem", "episodic.jsonl"),
		"skills",
		"logs",
	}
	for _, e := range entries {
		_, statErr := os.Stat(filepath.Join(dir, e))
		assert.NoError(t, statErr, "expected %s to exist", e)
	}
}

// Test 2: agentfs.yaml contains the provided name.
func TestInit_AgentfsYAMLContainsName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "testagent")
	name := "testagent"

	err := scaffold.Init(dir, name)
	require.NoError(t, err)

	data, readErr := os.ReadFile(filepath.Join(dir, "agentfs.yaml"))
	require.NoError(t, readErr)

	assert.Contains(t, string(data), name)
}

// Test 3: soul.md contains the provided name in both frontmatter and body.
func TestInit_SoulMDContainsName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "soultest")
	name := "soultest"

	err := scaffold.Init(dir, name)
	require.NoError(t, err)

	data, readErr := os.ReadFile(filepath.Join(dir, "soul.md"))
	require.NoError(t, readErr)

	content := string(data)
	// Name appears in frontmatter
	assert.True(t, strings.Contains(content, "name: "+name), "soul.md should contain name in frontmatter")
	// Name appears in body
	assert.True(t, strings.Contains(content, "You are "+name), "soul.md should contain name in body")
}

// Test 4: Init non-existent parent dir → creates parent dirs as needed.
func TestInit_CreatesParentDirs(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "deeply", "nested", "agent")

	err := scaffold.Init(dir, "nested-agent")
	require.NoError(t, err)

	info, statErr := os.Stat(dir)
	require.NoError(t, statErr)
	assert.True(t, info.IsDir())
}

// Test 5: Init existing non-empty directory → returns error containing "already exists".
func TestInit_ExistingNonEmptyDir_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	// Put a file in the directory so it's non-empty.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "somefile.txt"), []byte("data"), 0644))

	err := scaffold.Init(dir, "agent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// Test 6: episodic.jsonl is created as an empty file (0 bytes), not a directory.
func TestInit_EpisodicJSONLIsEmptyFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "filetest")

	err := scaffold.Init(dir, "filetest")
	require.NoError(t, err)

	episodicPath := filepath.Join(dir, "mem", "episodic.jsonl")
	info, statErr := os.Stat(episodicPath)
	require.NoError(t, statErr)

	assert.False(t, info.IsDir(), "episodic.jsonl should be a file, not a directory")
	assert.Equal(t, int64(0), info.Size(), "episodic.jsonl should be 0 bytes")
}
