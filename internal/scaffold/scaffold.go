package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

const agentfsYAML = `version: "1.0"
name: %s
description: ""
`

const soulMD = `---
name: %s
version: 1.0
---

# Identity
You are %s.

# Behavioral rules
- Be helpful and precise.
`

// Init creates a new agent directory at dir with the standard AgentFS layout.
// If dir already exists and is not empty, returns error.
// Creates:
//
//	dir/agentfs.yaml         (with provided name)
//	dir/soul.md              (template)
//	dir/mem/                 (empty directory)
//	dir/mem/episodic.jsonl   (empty file)
//	dir/skills/              (empty directory)
//	dir/logs/                (empty directory)
func Init(dir string, name string) error {
	// Check if dir already exists and is non-empty.
	info, err := os.Stat(dir)
	if err == nil && info.IsDir() {
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			return fmt.Errorf("reading directory %s: %w", dir, readErr)
		}
		if len(entries) > 0 {
			return fmt.Errorf("directory %s already exists and is not empty", dir)
		}
	}

	// Create the directory (and any parents) if it doesn't exist.
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Write agentfs.yaml
	yamlContent := fmt.Sprintf(agentfsYAML, name)
	if err := os.WriteFile(filepath.Join(dir, "agentfs.yaml"), []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("writing agentfs.yaml: %w", err)
	}

	// Write soul.md
	soulContent := fmt.Sprintf(soulMD, name, name)
	if err := os.WriteFile(filepath.Join(dir, "soul.md"), []byte(soulContent), 0644); err != nil {
		return fmt.Errorf("writing soul.md: %w", err)
	}

	// Create mem/ directory
	memDir := filepath.Join(dir, "mem")
	if err := os.MkdirAll(memDir, 0755); err != nil {
		return fmt.Errorf("creating mem directory: %w", err)
	}

	// Create mem/episodic.jsonl as empty file
	if err := os.WriteFile(filepath.Join(memDir, "episodic.jsonl"), []byte{}, 0644); err != nil {
		return fmt.Errorf("writing mem/episodic.jsonl: %w", err)
	}

	// Create skills/ directory
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0755); err != nil {
		return fmt.Errorf("creating skills directory: %w", err)
	}

	// Create logs/ directory
	if err := os.MkdirAll(filepath.Join(dir, "logs"), 0755); err != nil {
		return fmt.Errorf("creating logs directory: %w", err)
	}

	return nil
}
