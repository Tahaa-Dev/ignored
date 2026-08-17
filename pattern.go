//revive:disable:confusing-results
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
	// Determines if the given path matches the pattern, relative to the rootDir.
	// The first bool is whether the raw pattern is a match or not.
	// The second one is whether if it's ignored or not considering negation (leading !).
	Match(path string, isDir bool, rootDir string) (isMatch bool, result bool)
	// Checks if the already-normalized path matches the pattern.
	// The first bool is whether the raw pattern is a match or not.
	// The second one is whether if it's ignored or not considering negation (leading !).
	MatchNormalized(path string, isDir bool) (isMatch bool, result bool)
	// Returns any error encountered during pattern parsing.
	Err() error
}

// Function for constructing [Pattern] interfaces.
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
	if suffixRemoved := strings.TrimRight(pat, "/"); len(suffixRemoved) < len(pat) {
		pat = suffixRemoved
		isDir = true
	}

	if len(pat) == 0 {
		return &emptyPattern{}
	}

	if !strings.ContainsRune(pat, '/') {
		pat = "**/" + pat + "/**"
	} else {
		if !strings.HasPrefix(pat, "/") {
			pat = "/" + pat
		}

		if !strings.ContainsAny(pat, "*[]?\\{}") {
			if !strings.HasSuffix(pat, "/") {
				pat = pat + "/"
			}
			return &literalPattern{pat, isDir, isNeg}
		}
		if strings.HasSuffix(pat, "/") {
			pat = pat + "**"
		} else {
			pat = pat + "/**"
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

func (p *globPattern) Match(path string, isDir bool, rootDir string) (bool, bool) {
	path, err := NormalizePath(path, rootDir)
	if err != nil {
		return getRes(false, p.isNeg)
	}

	return p.MatchNormalized(path, isDir)
}

func (p *globPattern) MatchNormalized(path string, isDir bool) (bool, bool) {
	if p.isDir && !isDir {
		path = filepath.Dir(path)
	}

	return getRes(p.glob.Match(path), p.isNeg)
}

func (*globPattern) Err() error {
	return nil
}

type EmptyPatternErr struct{}

func (*EmptyPatternErr) Error() string { return "pattern is empty" }

type errPattern struct{ err error }

func (*errPattern) Match(_ string, _ bool, _ string) (bool, bool) { return false, false }
func (*errPattern) MatchNormalized(_ string, _ bool) (bool, bool) { return false, false }
func (p *errPattern) Err() error                                  { return p.err }

type emptyPattern struct{}

func (*emptyPattern) Match(_ string, _ bool, _ string) (bool, bool) { return false, false }
func (*emptyPattern) MatchNormalized(_ string, _ bool) (bool, bool) { return false, false }
func (*emptyPattern) Err() error                                    { return &EmptyPatternErr{} }

type literalPattern struct {
	literal string
	isDir   bool
	isNeg   bool
}

func (p *literalPattern) Match(path string, isDir bool, rootDir string) (bool, bool) {
	path, err := NormalizePath(path, rootDir)
	if err != nil {
		return getRes(false, p.isNeg)
	}

	return p.MatchNormalized(path, isDir)
}

func (p *literalPattern) MatchNormalized(path string, isDir bool) (bool, bool) {
	if !strings.HasPrefix(path, p.literal) {
		return getRes(false, p.isNeg)
	}

	var res bool
	if p.isDir && !isDir {
		res = len(path) > len(p.literal)
	} else {
		res = len(path) >= len(p.literal)
	}

	return getRes(res, p.isNeg)
}

func (*literalPattern) Err() error {
	return nil
}

// Helper function for normalizing paths used by [Pattern].Match() and [Matcher].Match().
func NormalizePath(path string, rootDir string) (string, error) {
	relPath, err := filepath.Rel(filepath.ToSlash(rootDir), filepath.ToSlash(path))
	if err != nil {
		return "", errors.New(
			"Path '" + path + "' is not related to root directory '" + rootDir + "'",
		)
	}
	if strings.HasSuffix(path, "/") {
		return "/" + relPath, nil
	}
	return "/" + relPath + "/", nil
}

func getRes(res bool, isNeg bool) (bool, bool) {
	if isNeg {
		return res, !res
	}
	return res, res
}
