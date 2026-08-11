package ignored

import (
	"errors"
	"path/filepath"
)

type Matcher interface {
	Match(path string, isDir bool) bool
	SetRootDir(rootDir string) Matcher
	Extend(patterns ...string) Matcher
	ExtendFromPatterns(patterns ...Pattern) Matcher
	Err() error
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
	m.err = errors.Join(m.err, err)

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
	for _, pat := range patterns {
		if p := ParsePattern(pat); p.Err() != nil {
			m.err = errors.Join(m.err, p.Err())
		} else {
			m.patterns = append(m.patterns, p)
		}
	}

	return m
}

func (m *matcher) ExtendFromPatterns(patterns ...Pattern) Matcher {
	for _, pat := range patterns {
		if pat.Err() != nil {
			m.err = errors.Join(m.err, pat.Err())
		} else {
			m.patterns = append(m.patterns, pat)
		}
	}

	return m
}

func (m *matcher) Err() error {
	return m.err
}
