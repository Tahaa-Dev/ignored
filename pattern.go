package ignored

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

type Pattern interface {
	Match(path string, isDir bool, rootDir string) (bool, error)
	MatchNormalized(path string, isDir bool) bool
	Err() error
}

func ParsePattern(pat string) Pattern {
	p := &globPattern{}

	if strings.HasPrefix(pat, "!") {
		pat = pat[1:]
		p.isNeg = true
	}

	if strings.HasSuffix(pat, "/") {
		pat = pat[:len(pat)-1]
		p.isDir = true
	}

	if len(pat) == 0 {
		return &errPattern{errors.New("Unexpected EOF: Pattern is empty")}
	}

	if !strings.ContainsRune(pat, '/') {
		pat = "**/" + pat
	} else {
		pat = strings.TrimLeft(pat, "/")
	}

	g, err := glob.Compile(pat, '/')
	if err != nil {
		return &errPattern{err}
	} else {
		p.glob = g
	}

	return p
}

type globPattern struct {
	glob  glob.Glob
	isDir bool
	isNeg bool
}

func (p *globPattern) Match(path string, isDir bool, rootDir string) (bool, error) {
	if p.isDir && !isDir {
		return false, nil
	}

	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return false, errors.New(
			"Path '" + path + "' is not related to root directory '" + rootDir + "'",
		)
	}
	path = filepath.ToSlash(relPath)

	return p.MatchNormalized(path, isDir), nil
}

func (p *globPattern) MatchNormalized(path string, isDir bool) bool {
	if p.isDir && !isDir {
		return false
	}

	match := p.glob.Match(path)
	if p.isNeg {
		return !match
	}
	return match
}

func (p *globPattern) Err() error {
	return nil
}

type errPattern struct {
	err error
}

func (p *errPattern) Match(path string, isDir bool, rootDir string) (bool, error) {
	return false, nil
}
func (p *errPattern) MatchNormalized(path string, isDir bool) bool {
	return false
}
func (p *errPattern) Err() error {
	return p.err
}
