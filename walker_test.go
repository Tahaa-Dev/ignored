package ignored

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestRepoWalker_WalkRepo(t *testing.T) {
	fsys := fstest.MapFS{
		".gitignore":        {Data: []byte("*.log")},
		"a.log":             {Data: []byte("ignore me")},
		"subdir":            {Mode: fs.ModeDir},
		"subdir/.gitignore": {Data: []byte("*.txt")},
		"subdir/b.log":      {Data: []byte("ignored via parent")},
		"subdir/c.txt":      {Data: []byte("ignored via local")},
		"subdir/d.go":       {Data: []byte("keep me")},
	}

	walker := NewRepoWalkerFS(fsys, ".")
	visited := make(map[string]bool)

	_ = walker.WalkRepo(func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		visited[path] = true
		return nil
	})

	if walker.Err() != nil {
		t.Fatalf("Walker error: %v", walker.Err())
	}

	expected := []string{
		".",
		"subdir",
		"subdir/d.go",
	}

	for _, p := range expected {
		if !visited[p] {
			t.Errorf("Expected path %q to be visited", p)
		}
	}

	unexpected := []string{
		"a.log",
		"subdir/b.log",
		"subdir/c.txt",
	}

	for _, p := range unexpected {
		if visited[p] {
			t.Errorf("Path %q should have been ignored", p)
		}
	}
}

func TestRepoWalker_SetIgnoreFileName(t *testing.T) {
	fsys := fstest.MapFS{
		".ignore": {Data: []byte("*.log")},
		"a.log":   {Data: []byte("ignore me")},
	}

	walker := NewRepoWalkerFS(fsys, ".")
	walker.SetIgnoreFileName(".ignore")

	visited := make(map[string]bool)
	_ = walker.WalkRepo(func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		visited[path] = true
		return nil
	})

	if visited["a.log"] {
		t.Error("a.log should have been ignored using .ignore file")
	}
}
