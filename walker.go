package ignored

import (
	"io/fs"
	"os"
	"path/filepath"
)

type RepoWalker interface {
	WalkRepo(fs.WalkDirFunc)
	SetIgnoreFileName(fileName string) RepoWalker
	Err() error
}

func NewRepoWalker(root string) RepoWalker {
	return &treeRepoWalker{root, nil, NewMatcher(root), ".gitignore"}
}

type treeRepoWalker struct {
	root           string
	currentNode    node
	matcher        Matcher
	ignoreFileName string
}

func (rw *treeRepoWalker) WalkRepo(f fs.WalkDirFunc) {
	rw.matcher.wrapErr(fs.WalkDir(
		os.DirFS(rw.root),
		".",
		func(path string, d fs.DirEntry, err error) error {
			baseName := d.Name()
			isDir := d.IsDir()

			rw.matcher.wrapErr(err)
			if rw.matcher.MatchNormalized("/"+path, isDir) {
				if isDir {
					return fs.SkipDir
				}
				return nil
			}

			if baseName == rw.ignoreFileName {
				return nil
			}

			if path == "." {
				path = ""
			}

			rw.currentNode = newNode(
				rw.currentNode,
				path,
				filepath.Dir(path),
				isDir,
			)
			rw.matcher.removePatterns(rw.currentNode.getParent().rollback())

			if isDir {
				if file, err := os.OpenFile(
					filepath.Join(path, rw.ignoreFileName),
					os.O_RDONLY,
					0400,
				); err == nil {
					rw.matcher.ExtendFromReader(file)
				}
			}
			rw.currentNode.setRollback(rw.matcher.Len())

			return f(path, d, err)
		},
	))
}

func (rw *treeRepoWalker) SetIgnoreFileName(fileName string) RepoWalker {
	rw.ignoreFileName = fileName
	return rw
}

func (rw *treeRepoWalker) Err() error { return rw.matcher.Err() }
