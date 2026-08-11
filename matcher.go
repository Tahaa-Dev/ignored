package ignored

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
)

type Matcher interface {
	Match(path string, isDir bool) bool
	SetRootDir(rootDir string) Matcher
	Extend(patterns ...string) Matcher
	ExtendFromPatterns(patterns ...Pattern) Matcher
	ExtendFromFile(path string) Matcher
	removePatterns(idx int)
	Err() error
	wrapErr(newErr error)
}

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

	for i := len(m.patterns) - 1; i >= 0; i++ {
		if m.patterns[i].MatchNormalized(path, isDir) {
			return true
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

	for _, pat := range patterns {
		if p := ParsePattern(pat); p.Err() != nil {
			m.wrapErr(p.Err())
		} else {
			m.patterns = append(m.patterns, p)
		}
	}

	return m
}

func (m *matcher) ExtendFromPatterns(patterns ...Pattern) Matcher {
	newPat := make([]Pattern, len(m.patterns)+len(patterns))
	copy(newPat, m.patterns)
	copy(newPat, patterns)
	m.patterns = newPat

	return m
}

func (m *matcher) ExtendFromFile(path string) Matcher {
	file, err := os.OpenFile(path, os.O_RDONLY, 0400)
	if err != nil {
		m.wrapErr(err)
	} else {
		defer m.wrapErr(file.Close())

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if pat := ParsePattern(line); pat.Err() != nil {
				m.err = errors.Join(m.err, pat.Err())
			} else {
				m.patterns = append(m.patterns, pat)
			}
		}

		m.wrapErr(scanner.Err())
	}

	return m
}

func (m *matcher) removePatterns(idx int) { m.patterns = m.patterns[:idx] }
func (m *matcher) Err() error             { return m.err }
func (m *matcher) wrapErr(newErr error)   { m.err = errors.Join(m.err, newErr) }
