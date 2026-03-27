package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a single skill loaded from a SKILL.md file.
type Skill struct {
	Name    string
	Content string
}

// Store manages the skills directory.
type Store struct {
	dir string
}

// Open opens (or creates) the skills directory at dir.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("skills: open %q: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// List returns all skills in the store, sorted by filename.
func (s *Store) List() ([]Skill, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("skills: list: %w", err)
	}

	var skills []Skill
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		skill, err := s.load(name)
		if err != nil {
			return nil, err
		}
		skills = append(skills, *skill)
	}
	return skills, nil
}

// Get returns the skill with the given name (filename without .md extension).
// Returns an error if the skill does not exist.
func (s *Store) Get(name string) (*Skill, error) {
	filename := name
	if !strings.HasSuffix(filename, ".md") {
		filename = name + ".md"
	}
	return s.load(filename)
}

func (s *Store) load(filename string) (*Skill, error) {
	path := filepath.Join(s.dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("skills: load %q: %w", filename, err)
	}
	name := strings.TrimSuffix(filename, ".md")
	return &Skill{Name: name, Content: string(data)}, nil
}
