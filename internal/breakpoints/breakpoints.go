package breakpoints

import "fmt"

// Breakpoint represents a line breakpoint set in Xdebug.
type Breakpoint struct {
	File string // short filename, e.g. "index.php"
	Line int
	ID   string // Xdebug-assigned ID from breakpoint_set response
}

// Store is an in-memory collection of active breakpoints.
type Store struct {
	items []Breakpoint
}

// Add adds a breakpoint to the store.
func (s *Store) Add(file string, line int, id string) {
	s.items = append(s.items, Breakpoint{File: file, Line: line, ID: id})
}

// Remove removes the breakpoint at the given file:line.
// Returns the Xdebug ID of the removed breakpoint, or an error if not found.
func (s *Store) Remove(file string, line int) (id string, err error) {
	for i, bp := range s.items {
		if bp.File == file && bp.Line == line {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return bp.ID, nil
		}
	}
	return "", fmt.Errorf("no breakpoint at %s:%d", file, line)
}

// Format returns a display string for the Breakpoints panel, one entry per line.
// Returns "(none)" if empty.
func (s *Store) Format() string {
	if len(s.items) == 0 {
		return "(none)"
	}
	result := ""
	for _, bp := range s.items {
		result += fmt.Sprintf("%s:%d\n", bp.File, bp.Line)
	}
	return result
}
