package skills_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aaron/agent-fuse/internal/skills"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempStore(t *testing.T) *skills.Store {
	t.Helper()
	store, err := skills.Open(t.TempDir())
	require.NoError(t, err)
	return store
}

func writeSkill(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".md"), []byte(content), 0644))
}

func TestOpen_CreatesDirectoryIfNotExist(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "skills")

	_, err := os.Stat(dir)
	require.True(t, os.IsNotExist(err))

	store, err := skills.Open(dir)
	require.NoError(t, err)
	require.NotNil(t, store)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	list, err := store.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

// List returns skills from .md files in the directory
func TestList_ReturnsMDFiles(t *testing.T) {
	dir := t.TempDir()
	store, err := skills.Open(dir)
	require.NoError(t, err)

	writeSkill(t, dir, "web-search", "# Web Search\nSearches the web.")
	writeSkill(t, dir, "code-review", "# Code Review\nReviews code.")

	list, err := store.List()
	require.NoError(t, err)
	require.Len(t, list, 2)

	names := map[string]bool{list[0].Name: true, list[1].Name: true}
	assert.True(t, names["web-search"])
	assert.True(t, names["code-review"])
}

// List ignores non-.md files
func TestList_IgnoresNonMDFiles(t *testing.T) {
	dir := t.TempDir()
	store, err := skills.Open(dir)
	require.NoError(t, err)

	writeSkill(t, dir, "valid", "# Valid Skill")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("not a skill"), 0644))

	list, err := store.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "valid", list[0].Name)
}

// Get returns skill content by name
func TestGet_ReturnsSkillContent(t *testing.T) {
	dir := t.TempDir()
	store, err := skills.Open(dir)
	require.NoError(t, err)

	writeSkill(t, dir, "my-skill", "# My Skill\nDoes something useful.")

	skill, err := store.Get("my-skill")
	require.NoError(t, err)
	assert.Equal(t, "my-skill", skill.Name)
	assert.Equal(t, "# My Skill\nDoes something useful.", skill.Content)
}

// Get returns error for missing skill
func TestGet_MissingSkill(t *testing.T) {
	store := tempStore(t)

	_, err := store.Get("nonexistent")
	assert.Error(t, err)
}
