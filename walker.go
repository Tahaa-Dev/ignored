package ignored

import (
	"io/fs"
	"os"
	"path/filepath"
)

// RepoWalker provides an ignore file compliant wrapper for [fs.WalkDir].
// It is constructed via the [NewRepoWalker] function.
//
// ## Example:
//
//	walker := ignored.NewRepoWalker("/home/user/project")
//	walker.WalkRepo(func(path string, d fs.DirEntry, err error) error {
//		if err != nil {
//			return err
//		}
//		fmt.Println("Visiting:", path)
//		return nil
//	})
//	if walker.Err() != nil {
//		// handle error
//	}
type RepoWalker interface {
	// WalkRepo traverses the repository starting at the root, applying the [Matcher] to skip ignored files and directories. It uses the provided [fs.WalkDirFunc].
	WalkRepo(fs.WalkDirFunc)
	// SetIgnoreFileName changes the name of the file used to detect ignore patterns (default is ".gitignore").
	SetIgnoreFileName(fileName string) RepoWalker
	// Err returns any error that occurred during the repository traversal.
	Err() error
}

// Function for constructing [RepoWalker] interfaces
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
