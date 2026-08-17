package ignored

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
	WalkRepo(fs.WalkDirFunc) error
	// Changes the name of the file used to detect ignore patterns (default is ".gitignore").
	SetIgnoreFileName(fileName string) RepoWalker
	// Returns any error that occurred during the repository traversal.
	Err() error
}

// Function for constructing [RepoWalker] interfaces.
func NewRepoWalker(root string, patterns ...string) RepoWalker {
	return &treeRepoWalker{os.DirFS(root), nil, NewMatcher("", patterns...), ".gitignore"}
}

// Function for constructing [RepoWalker] interfaces from an [fs.FS].
func NewRepoWalkerFS(rootFS fs.FS, patterns ...string) RepoWalker {
	return &treeRepoWalker{rootFS, nil, NewMatcher("", patterns...), ".gitignore"}
}

type treeRepoWalker struct {
	root           fs.FS
	currentNode    node
	matcher        Matcher
	ignoreFileName string
}

func (rw *treeRepoWalker) WalkRepo(f fs.WalkDirFunc) error {
	return fs.WalkDir(
		rw.root,
		".",
		func(path string, d fs.DirEntry, err error) error {
			baseName := d.Name()
			isDir := d.IsDir()

			rw.matcher.wrapErr(err)

			if path != "." && rw.matcher.Match(path, isDir) {
				if isDir {
					return fs.SkipDir
				}
				return nil
			}

			if baseName == rw.ignoreFileName {
				return nil
			}

			if path == "." {
				rw.currentNode = &dirNode{".", nil, nil, 0}
			} else {
				trimmedPath := strings.TrimRight(path, "/")
				rw.currentNode = newNode(
					rw.currentNode,
					trimmedPath,
					filepath.Dir(trimmedPath),
					isDir,
				)
				rw.matcher.removePatterns(rw.currentNode.getParent().rollback())
			}

			// Uses manual Matcher.ExtendFromReader instead of Matcher.ExtendFromFile so Err doesn't
			// get cluttered by failed file open attempts
			if isDir {
				if file, err := rw.root.Open(
					filepath.Join(path, rw.ignoreFileName),
				); err == nil {
					rw.matcher.ExtendFromReader(file)
					rw.matcher.wrapErr(file.Close())
				}
			}
			rw.currentNode.setRollback(rw.matcher.Len())

			return f(path, d, err)
		},
	)
}

func (rw *treeRepoWalker) SetIgnoreFileName(fileName string) RepoWalker {
	rw.ignoreFileName = fileName
	return rw
}

func (rw *treeRepoWalker) Err() error { return rw.matcher.Err() }
