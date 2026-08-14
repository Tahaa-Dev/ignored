package ignored

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

// Pattern provides methods to handle ignore patterns individually.
// It is parsed from a string via the [ParsePattern] function.
//
// Example:
//
//	pat := ignored.ParsePattern("*.log")
//	isIgnored, err := pat.Match("app.log", false, "/home/user/project")
//	if err != nil {
//		// handle error
//	}
//	if isIgnored {
//		fmt.Println("File is ignored")
//	}
type Pattern interface {
	// Match determines if the given path matches the pattern, relative to the rootDir.
	// It returns true if the path is matched, and an error if normalization fails.
	Match(path string, isDir bool, rootDir string) (bool, error)
	// MatchNormalized checks if the already-normalized path matches the pattern.
	MatchNormalized(path string, isDir bool) bool
	// Err returns any error encountered during pattern parsing or matching.
	Err() error
}

// Function for constructing [Pattern] interfaces
func ParsePattern(pat string) Pattern {
	isNeg := false
	isDir := false

	if idx := strings.IndexRune(pat, '#'); idx != -1 {
		pat = pat[:idx]
	}
	pat = strings.TrimSpace(pat)

	if strings.HasPrefix(pat, "!") {
		pat = pat[1:]
		isNeg = true
	}

	if strings.HasSuffix(pat, "/") {
		pat = pat[:len(pat)-1]
		isDir = true
	}

	if len(pat) == 0 {
		return &errPattern{errors.New("unexpected EOF: pattern is empty")}
	}

	if !strings.ContainsRune(pat, '/') {
		if !strings.ContainsAny(pat, "*[]?\\{}") {
			return &baseNamePattern{pat, isDir, isNeg}
		}
		pat = "**/" + pat
	} else if pat[0] != '/' {
		pat = "/" + pat
		if !strings.ContainsAny(pat, "*[]?\\{}") {
			return &literalPattern{pat, isDir, isNeg}
		}
	}

	g, err := glob.Compile(pat, '/')
	if err != nil {
		return &errPattern{err}
	}

	return &globPattern{g, isDir, isNeg}
}

type globPattern struct {
	glob  glob.Glob
	isDir bool
	isNeg bool
}

func (p *globPattern) Match(path string, isDir bool, rootDir string) (bool, error) {
	if p.isDir && !isDir {
		return getRes(false, p.isNeg), nil
	}

	path, err := NormalizePath(path, rootDir)
	if err != nil {
		return getRes(false, p.isNeg), err
	}

	return p.MatchNormalized(path, isDir), nil
}

func (p *globPattern) MatchNormalized(path string, isDir bool) bool {
	if p.isDir && !isDir {
		return getRes(false, p.isNeg)
	}

	match := p.glob.Match(path)
	return getRes(match, p.isNeg)
}

func (*globPattern) Err() error {
	return nil
}

type errPattern struct {
	err error
}

func (*errPattern) Match(_ string, _ bool, _ string) (bool, error) {
	return false, nil
}
func (*errPattern) MatchNormalized(_ string, _ bool) bool {
	return false
}
func (p *errPattern) Err() error {
	return p.err
}

type literalPattern struct {
	literal string
	isDir   bool
	isNeg   bool
}

func (p *literalPattern) Match(path string, isDir bool, rootDir string) (bool, error) {
	path, err := NormalizePath(path, rootDir)
	if err != nil {
		return getRes(false, p.isNeg), err
	}

	return p.MatchNormalized(path, isDir), nil
}

func (p *literalPattern) MatchNormalized(path string, isDir bool) bool {
	if !strings.HasPrefix(path, p.literal) {
		return getRes(false, p.isNeg)
	}

	if p.isDir {
		return getRes(
			(isDir && len(path) == len(p.literal)) ||
				(len(path) > len(p.literal)+1 && path[len(p.literal)] == '/'),
			p.isNeg,
		)
	}

	return getRes(len(path) == len(p.literal), p.isNeg)
}

func (*literalPattern) Err() error {
	return nil
}

type baseNamePattern struct {
	baseName string
	isDir    bool
	isNeg    bool
}

func (p *baseNamePattern) Match(path string, isDir bool, _ string) (bool, error) {
	return p.MatchNormalized(path, isDir), nil
}

func (p *baseNamePattern) MatchNormalized(path string, isDir bool) bool {
	return filepath.Base(filepath.ToSlash(path)) == p.baseName && p.isDir == isDir
}

func (*baseNamePattern) Err() error {
	return nil
}

// Helper function for normalizing paths used by [Pattern].Match() and [Matcher].Match()
func NormalizePath(path string, rootDir string) (string, error) {
	relPath, err := filepath.Rel(filepath.ToSlash(rootDir), filepath.ToSlash(path))
	if err != nil {
		return "", errors.New(
			"Path '" + path + "' is not related to root directory '" + rootDir + "'",
		)
	}
	return "/" + strings.TrimRight(relPath, "/"), nil
}

func getRes(res bool, isNeg bool) bool {
	if isNeg {
		return !res
	}
	return res
}
