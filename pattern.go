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
	p := &pattern{}

	if strings.HasPrefix(pat, "!") {
		pat = pat[1:]
		p.isNeg = true
	}

	if strings.HasSuffix(pat, "/") {
		pat = pat[:len(pat)-1]
		p.isDir = true
	}

	if len(pat) == 0 {
		p.err = errors.New("Unexpected EOF: Pattern is empty")
		return p
	}

	if !strings.ContainsRune(pat, '/') {
		pat = "**/" + pat
	} else {
		pat = strings.TrimLeft(pat, "/")
	}

	g, err := glob.Compile(pat, '/')
	if err != nil {
		p.err = err
	} else {
		p.glob = g
	}

	return p
}

type pattern struct {
	glob  glob.Glob
	isDir bool
	isNeg bool
	err   error
}

func (p *pattern) Match(path string, isDir bool, rootDir string) (bool, error) {
	if p.err != nil || (p.isDir && !isDir) {
		return false, nil
	}

	relPath, err := filepath.Rel(rootDir, path)
	if err != nil {
		return false, err
	}
	path = filepath.ToSlash(relPath)

	return p.MatchNormalized(path, isDir), nil
}

func (p *pattern) MatchNormalized(path string, isDir bool) bool {
	if p.err != nil || (p.isDir && !isDir) {
		return false
	}

	match := p.glob.Match(path)
	if p.isNeg {
		return !match
	}
	return match
}

func (p *pattern) Err() error {
	return p.err
}
