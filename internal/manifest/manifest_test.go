package manifest_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aaron/agent-fuse/internal/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: write a temp agentfs.yaml with given content and return path
func writeTempManifest(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "agentfs.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

// Test 1: Parse a valid agentfs.yaml → returns Manifest with correct fields
func TestParse_ValidManifest(t *testing.T) {
	path := writeTempManifest(t, `
version: "1.0"
name: ResearchAgent
description: A research assistant for technical due diligence
`)
	m, err := manifest.Parse(path)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, "1.0", m.Version)
	assert.Equal(t, "ResearchAgent", m.Name)
	assert.Equal(t, "A research assistant for technical due diligence", m.Description)
}

// Test 2: Parse agentfs.yaml with missing Name → error containing "missing required field: name"
func TestParse_MissingName(t *testing.T) {
	path := writeTempManifest(t, `
version: "1.0"
description: No name here
`)
	m, err := manifest.Parse(path)
	assert.Nil(t, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required field: name")
}

// Test 3: Parse a non-existent file → returns error wrapping os.ErrNotExist
func TestParse_FileNotFound(t *testing.T) {
	_, err := manifest.Parse("/nonexistent/path/agentfs.yaml")
	require.Error(t, err)
	assert.True(t, errors.Is(err, os.ErrNotExist), "expected os.ErrNotExist, got: %v", err)
}

// Test 4: Parse invalid YAML → returns error
func TestParse_InvalidYAML(t *testing.T) {
	path := writeTempManifest(t, `
name: [unclosed bracket
  bad: yaml: here
`)
	_, err := manifest.Parse(path)
	require.Error(t, err)
}

// Test 5: Parse agentfs.yaml with no Version → Version defaults to "1.0"
func TestParse_DefaultVersion(t *testing.T) {
	path := writeTempManifest(t, `
name: MinimalAgent
`)
	m, err := manifest.Parse(path)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, "1.0", m.Version)
	assert.Equal(t, "MinimalAgent", m.Name)
}

// Test 6: Parse agentfs.yaml with extra/unknown fields → succeeds (forward-compatibility)
func TestParse_UnknownFields(t *testing.T) {
	path := writeTempManifest(t, `
version: "1.0"
name: FutureAgent
description: Has extra fields
future_field: some_value
another_unknown: 42
`)
	m, err := manifest.Parse(path)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, "FutureAgent", m.Name)
	assert.Equal(t, "1.0", m.Version)
}
