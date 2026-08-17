package ignored

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Matcher provides methods to handle multiple ignore patterns at once.
// It is constructed via the [NewMatcher] function.
//
// ## Example:
//
//	matcher := ignored.NewMatcher("/home/user/project", "*.log", "node_modules/")
//	if matcher.Match("node_modules/pkg/main.js", false) {
//		fmt.Println("Path is ignored")
//	}
//
//	// Dynamically add patterns
//	matcher.Extend("*.tmp")
//	matcher.ExtendFromFile("/home/user/project/.gitignore")
type Matcher interface {
	// Matches path against [Pattern]s in the [Matcher].
	Match(path string, isDir bool) bool
	// Matches already normalized path against [Pattern]s in the [Matcher].
	MatchNormalized(path string, isDir bool) bool
	// Udates the root directory for the [Matcher].
	SetRootDir(rootDir string) Matcher
	// Appends new patterns from a slice of strings to the [Matcher].
	Extend(patterns ...string) Matcher
	// Appends new pre-parsed [Pattern]s to the [Matcher].
	ExtendFromPatterns(patterns ...Pattern) Matcher
	// Loads patterns from a file at the given path into the [Matcher].
	ExtendFromFile(path string) Matcher
	// Loads patterns from an [io.Reader] into the [Matcher].
	ExtendFromReader(reader io.Reader) Matcher
	// Rolls back the [Matcher]'s patterns to the specified index.
	removePatterns(idx int)
	// Returns the first accumulated error in the [Matcher].
	Err() error
	// Returns the number of patterns currently in the [Matcher].
	Len() int
	// Accumulates a new error into the [Matcher]'s existing error.
	wrapErr(newErr error)
}

// Function for constructing [Matcher] interfaces.
func NewMatcher(rootDir string, patterns ...string) Matcher {
	return (&matcher{patterns: make([]Pattern, 0, 7)}).SetRootDir(rootDir).Extend(patterns...)
}

type matcher struct {
	patterns []Pattern
	rootDir  string
	err      error
}

func (m *matcher) Match(path string, isDir bool) bool {
	path, err := NormalizePath(path, m.rootDir)
	m.wrapErr(err)

	return m.MatchNormalized(path, isDir)
}

func (m *matcher) MatchNormalized(path string, isDir bool) bool {
	for i := len(m.patterns) - 1; i >= 0; i-- {
		if isMatch, res := m.patterns[i].MatchNormalized(path, isDir); isMatch {
			return res
		}
	}

	return false
}

func (m *matcher) SetRootDir(rootDir string) Matcher {
	m.rootDir = filepath.ToSlash(rootDir)
	return m
}

func (m *matcher) Extend(patterns ...string) Matcher {
	newPat := make([]Pattern, len(m.patterns), len(m.patterns)+len(patterns))
	copy(newPat, m.patterns)
	m.patterns = newPat

	for _, pat := range patterns {
		p := ParsePattern(pat)
		if err := p.Err(); err != nil {
			if !errors.Is(err, &EmptyPatternErr{}) {
				m.wrapErr(err)
			}
		} else {
			m.patterns = append(m.patterns, p)
		}
	}

	return m
}

func (m *matcher) ExtendFromPatterns(patterns ...Pattern) Matcher {
	newPat := make([]Pattern, len(m.patterns)+len(patterns))
	copy(newPat, m.patterns)
	copy(newPat[len(m.patterns):], patterns)
	m.patterns = newPat

	return m
}

func (m *matcher) ExtendFromFile(path string) Matcher {
	file, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		m.wrapErr(err)
	} else {
		defer func() {
			m.wrapErr(file.Close())
		}()
		m.ExtendFromReader(file)
	}

	return m
}

func (m *matcher) ExtendFromReader(reader io.Reader) Matcher {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		pat := ParsePattern(line)
		if err := pat.Err(); err != nil {
			if !errors.Is(err, &EmptyPatternErr{}) {
				m.wrapErr(err)
			}
		} else {
			m.patterns = append(m.patterns, pat)
		}
	}

	m.wrapErr(scanner.Err())

	return m
}

func (m *matcher) removePatterns(idx int) { m.patterns = m.patterns[:idx] }
func (m *matcher) Err() error             { return m.err }
func (m *matcher) wrapErr(newErr error)   { m.err = errors.Join(m.err, newErr) }
func (m *matcher) Len() int               { return len(m.patterns) }
