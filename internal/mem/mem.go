package mem

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry is a single episodic memory record.
type Entry struct {
	Timestamp time.Time `json:"ts"`
	Text      string    `json:"text"`
	Tags      []string  `json:"tags,omitempty"`
}

// EpisodicLog manages an append-only JSONL memory file.
type EpisodicLog struct {
	mu   sync.Mutex
	file *os.File
	path string
}

// Open opens (or creates) the episodic log at path.
// On open, it truncates to the last complete newline (crash recovery).
func Open(path string) (*EpisodicLog, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	// Crash recovery: truncate to last complete line.
	if err := truncateToLastNewline(f); err != nil {
		f.Close()
		return nil, err
	}

	// Seek to end for appending.
	if _, err := f.Seek(0, 2); err != nil {
		f.Close()
		return nil, err
	}

	return &EpisodicLog{file: f, path: path}, nil
}

// truncateToLastNewline truncates the file to the byte offset of the last '\n'.
// If the file ends with '\n' (or is empty), nothing changes.
func truncateToLastNewline(f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}
	size := info.Size()
	if size == 0 {
		return nil
	}

	// Read last byte to check if file ends with newline.
	buf := make([]byte, 1)
	if _, err := f.ReadAt(buf, size-1); err != nil {
		return err
	}
	if buf[0] == '\n' {
		// File ends cleanly; nothing to do.
		return nil
	}

	// Find the last newline by scanning backwards.
	const chunkSize = 4096
	lastNL := int64(-1)
	for off := size - 1; off >= 0; off -= chunkSize {
		start := off - chunkSize + 1
		if start < 0 {
			start = 0
		}
		chunk := make([]byte, off-start+1)
		if _, err := f.ReadAt(chunk, start); err != nil {
			return err
		}
		for i := len(chunk) - 1; i >= 0; i-- {
			if chunk[i] == '\n' {
				lastNL = start + int64(i)
				break
			}
		}
		if lastNL >= 0 {
			break
		}
		if start == 0 {
			break
		}
	}

	// Truncate to just after the last newline (or to 0 if none found).
	newSize := lastNL + 1
	if lastNL < 0 {
		newSize = 0
	}
	return f.Truncate(newSize)
}

// Append writes one entry as a single JSON line + newline.
// Safe for concurrent callers.
func (e *EpisodicLog) Append(entry Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	e.mu.Lock()
	defer e.mu.Unlock()

	_, err = e.file.Write(data)
	return err
}

// ReadAll reads all entries from the log and returns them in order.
// Returns an empty slice (not an error) if the file is empty.
func (e *EpisodicLog) ReadAll() ([]Entry, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, err := e.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var entries []Entry
	scanner := bufio.NewScanner(e.file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var entry Entry
		if err := json.Unmarshal(line, &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if entries == nil {
		entries = []Entry{}
	}
	return entries, nil
}

// Close releases any held resources.
func (e *EpisodicLog) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.file.Close()
}
