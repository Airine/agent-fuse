package logs_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aaron/agent-fuse/internal/logs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempLog(t *testing.T) (string, *logs.Logger) {
	t.Helper()
	path := fmt.Sprintf("%s/logs-test-%d.jsonl", t.TempDir(), time.Now().UnixNano())
	logger, err := logs.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { logger.Close() })
	return path, logger
}

// 1. Open a new log → creates file, ReadAll returns empty slice
func TestOpen_NewFile(t *testing.T) {
	_, logger := tempLog(t)
	events, err := logger.ReadAll()
	require.NoError(t, err)
	assert.Empty(t, events)
}

// 2. Log one event → ReadAll returns that event with correct fields + timestamp
func TestLog_OneEvent(t *testing.T) {
	_, logger := tempLog(t)

	err := logger.Log(logs.Event{Op: "read", Path: "/agentfs/soul.md"})
	require.NoError(t, err)

	events, err := logger.ReadAll()
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "read", events[0].Op)
	assert.Equal(t, "/agentfs/soul.md", events[0].Path)
	assert.False(t, events[0].Timestamp.IsZero(), "timestamp should be set automatically")
}

// 3. Log multiple events → ReadAll returns all in order
func TestLog_MultipleEvents(t *testing.T) {
	_, logger := tempLog(t)

	ops := []string{"read", "write", "exec"}
	for _, op := range ops {
		require.NoError(t, logger.Log(logs.Event{Op: op, Path: "/agentfs/" + op}))
	}

	events, err := logger.ReadAll()
	require.NoError(t, err)
	require.Len(t, events, 3)
	for i, op := range ops {
		assert.Equal(t, op, events[i].Op)
	}
}

// 4. Concurrent writes: 10 goroutines × 10 events = 100 total, all valid JSON
func TestLog_Concurrent(t *testing.T) {
	_, logger := tempLog(t)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				err := logger.Log(logs.Event{
					Op:        "read",
					Path:      fmt.Sprintf("/agentfs/worker-%d-%d", n, j),
					SessionID: fmt.Sprintf("s%d", n),
				})
				assert.NoError(t, err)
			}
		}(i)
	}
	wg.Wait()

	events, err := logger.ReadAll()
	require.NoError(t, err)
	assert.Len(t, events, 100)
}

// 5. Persistence: open existing log → ReadAll returns previously written events
func TestOpen_Persistence(t *testing.T) {
	path, logger := tempLog(t)
	require.NoError(t, logger.Log(logs.Event{Op: "write", Path: "/agentfs/mem/episodic.jsonl"}))
	require.NoError(t, logger.Close())

	logger2, err := logs.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { logger2.Close() })

	events, err := logger2.ReadAll()
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "write", events[0].Op)
}

// 6. Empty SessionID → omitted from JSON (omitempty), field is empty string on read
func TestLog_EmptySessionIDOmitted(t *testing.T) {
	_, logger := tempLog(t)
	require.NoError(t, logger.Log(logs.Event{Op: "read", Path: "/agentfs/soul.md"}))

	events, err := logger.ReadAll()
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Empty(t, events[0].SessionID)
}
