package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Event represents a single access event logged by an agent.
type Event struct {
	Timestamp time.Time `json:"ts"`
	Op        string    `json:"op"`
	Path      string    `json:"path"`
	SessionID string    `json:"session_id,omitempty"`
}

// Logger is a read-write append log for agent access events.
type Logger struct {
	mu   sync.Mutex
	path string
	f    *os.File
}

// Open opens (or creates) the log file at path for appending.
func Open(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("logs: open %q: %w", path, err)
	}
	return &Logger{path: path, f: f}, nil
}

// Log appends one Event as a JSON line. Safe for concurrent callers.
func (l *Logger) Log(event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	line, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("logs: marshal event: %w", err)
	}
	line = append(line, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = l.f.Write(line)
	return err
}

// ReadAll reads all events from the log file, returns them in order.
func (l *Logger) ReadAll() ([]Event, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("logs: seek: %w", err)
	}

	var events []Event
	scanner := bufio.NewScanner(l.f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("logs: parse entry: %w", err)
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("logs: scan: %w", err)
	}
	return events, nil
}

// Close flushes and closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.f.Close()
}
