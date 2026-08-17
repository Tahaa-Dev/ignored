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
	// Traverses the repository starting at the root, applying the [Matcher] to skip ignored files and directories.
	// Uses the provided [fs.WalkDirFunc].
	WalkRepo(fs.WalkDirFunc)
	// Changes the name of the file used to detect ignore patterns (default is ".gitignore").
	SetIgnoreFileName(fileName string) RepoWalker
	// Returns any error that occurred during the repository traversal.
	Err() error
}

// Function for constructing [RepoWalker] interfaces.
func NewRepoWalker(root string, patterns ...string) RepoWalker {
	return &treeRepoWalker{os.DirFS(root), nil, NewMatcher(root, patterns...), ".gitignore"}
}

// Function for constructing [RepoWalker] interfaces from an [fs.FS].
func NewRepoWalkerFS(rootFS fs.FS, root string, patterns ...string) RepoWalker {
	return &treeRepoWalker{rootFS, nil, NewMatcher(root, patterns...), ".gitignore"}
}

type treeRepoWalker struct {
	root           fs.FS
	currentNode    node
	matcher        Matcher
	ignoreFileName string
}

func (rw *treeRepoWalker) WalkRepo(f fs.WalkDirFunc) {
	rw.matcher.wrapErr(fs.WalkDir(
		rw.root,
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

			// Uses manual Matcher.ExtendFromReader instead of Matcher.ExtendFromFile so Err doesn't
			// get cluttered by failed file open attempts
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
