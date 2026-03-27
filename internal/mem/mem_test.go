package mem_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aaron/agent-fuse/internal/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test 1: Open a new (non-existent) log path → creates the file, ReadAll returns []
func TestOpen_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "episodic.jsonl")

	log, err := mem.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { log.Close() })

	// File should have been created.
	_, statErr := os.Stat(path)
	assert.NoError(t, statErr, "file should exist after Open")

	entries, err := log.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, entries, "ReadAll on new file should return empty slice")
}

// Test 2: Append one entry → ReadAll returns that entry with correct Text and timestamp.
func TestAppend_OneEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "episodic.jsonl")

	log, err := mem.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { log.Close() })

	now := time.Now().UTC().Truncate(time.Second)
	entry := mem.Entry{
		Timestamp: now,
		Text:      "hello world",
		Tags:      []string{"greet"},
	}

	err = log.Append(entry)
	require.NoError(t, err)

	entries, err := log.ReadAll()
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "hello world", entries[0].Text)
	assert.Equal(t, now, entries[0].Timestamp.UTC().Truncate(time.Second))
	assert.Equal(t, []string{"greet"}, entries[0].Tags)
}

// Test 3: Append multiple entries → ReadAll returns all in order.
func TestAppend_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "episodic.jsonl")

	log, err := mem.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { log.Close() })

	texts := []string{"first", "second", "third", "fourth", "fifth"}
	for i, text := range texts {
		err = log.Append(mem.Entry{
			Timestamp: time.Now().UTC(),
			Text:      text,
			Tags:      []string{fmt.Sprintf("tag%d", i)},
		})
		require.NoError(t, err)
	}

	entries, err := log.ReadAll()
	require.NoError(t, err)
	require.Len(t, entries, len(texts))
	for i, e := range entries {
		assert.Equal(t, texts[i], e.Text)
	}
}

// Test 4: Crash recovery: Open a file with a partial last line (no trailing newline)
// → truncates to last complete entry, ReadAll returns only complete entries.
func TestOpen_CrashRecovery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "episodic.jsonl")

	// Write two complete lines plus a partial third line (simulating crash).
	complete := `{"ts":"2024-01-01T00:00:00Z","text":"line one"}` + "\n" +
		`{"ts":"2024-01-01T00:00:01Z","text":"line two"}` + "\n" +
		`{"ts":"2024-01-01T00:00:02Z","text":"partial`
	err := os.WriteFile(path, []byte(complete), 0o644)
	require.NoError(t, err)

	log, err := mem.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { log.Close() })

	entries, err := log.ReadAll()
	require.NoError(t, err)
	require.Len(t, entries, 2, "should only have the two complete entries")
	assert.Equal(t, "line one", entries[0].Text)
	assert.Equal(t, "line two", entries[1].Text)
}

// Test 5: Concurrent appends: 10 goroutines each append 10 entries simultaneously
// → ReadAll returns exactly 100 entries, each entry is valid (no interleaved bytes).
func TestAppend_Concurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "episodic.jsonl")

	log, err := mem.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { log.Close() })

	const goroutines = 10
	const perGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		g := g
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				entry := mem.Entry{
					Timestamp: time.Now().UTC(),
					Text:      fmt.Sprintf("goroutine-%d-entry-%d", g, i),
				}
				err := log.Append(entry)
				assert.NoError(t, err)
			}
		}()
	}
	wg.Wait()

	entries, err := log.ReadAll()
	require.NoError(t, err)
	assert.Len(t, entries, goroutines*perGoroutine, "should have exactly 100 entries")

	// Verify every entry has non-empty Text (no corruption).
	for i, e := range entries {
		assert.NotEmpty(t, e.Text, "entry %d should have non-empty text", i)
		assert.False(t, e.Timestamp.IsZero(), "entry %d should have non-zero timestamp", i)
	}
}

// Test 6: Open a file that exists with valid entries → ReadAll returns existing entries
// (persistence across open/close).
func TestOpen_Persistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "episodic.jsonl")

	// First session: write some entries.
	{
		log, err := mem.Open(path)
		require.NoError(t, err)

		for _, text := range []string{"alpha", "beta", "gamma"} {
			require.NoError(t, log.Append(mem.Entry{
				Timestamp: time.Now().UTC(),
				Text:      text,
			}))
		}
		require.NoError(t, log.Close())
	}

	// Second session: re-open and verify entries are still there.
	{
		log, err := mem.Open(path)
		require.NoError(t, err)
		t.Cleanup(func() { log.Close() })

		entries, err := log.ReadAll()
		require.NoError(t, err)
		require.Len(t, entries, 3)
		assert.Equal(t, "alpha", entries[0].Text)
		assert.Equal(t, "beta", entries[1].Text)
		assert.Equal(t, "gamma", entries[2].Text)
	}
}
