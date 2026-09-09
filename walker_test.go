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

func TestRepoWalker_EdgeCases(t *testing.T) {
	t.Run("Empty FS", func(t *testing.T) {
		fsys := fstest.MapFS{
			".": {Mode: fs.ModeDir},
		}

		walker := NewRepoWalkerFS(fsys)
		visitedCount := 0
		err := walker.WalkRepo(func(_ string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			visitedCount++
			return nil
		})
		if err != nil {
			t.Fatalf("WalkRepo failed: %v", err)
		}
		if visitedCount != 1 {
			t.Errorf("Expected 1 path visited (.), got %d", visitedCount)
		}
	})

	t.Run("No Ignore File", func(t *testing.T) {
		fsys := fstest.MapFS{
			".":     {Mode: fs.ModeDir},
			"a.txt": {Data: []byte("content")},
			"b.log": {Data: []byte("content")},
		}
		walker := NewRepoWalkerFS(fsys)
		visited := make(map[string]bool)
		_ = walker.WalkRepo(func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			visited[path] = true
			return nil
		})

		expected := []string{".", "a.txt", "b.log"}
		for _, p := range expected {
			if !visited[p] {
				t.Errorf("Expected path %q to be visited", p)
			}
		}
	})

	t.Run("Ignore Scope Nested Dir", func(t *testing.T) {
		fsys := fstest.MapFS{
			".":                                   {Mode: fs.ModeDir},
			".gitignore":                          {Data: []byte("ignored_dir")},
			"ignored_dir":                         {Mode: fs.ModeDir},
			"ignored_dir/a.txt":                   {Data: []byte("content")},
			"nested_dir":                          {Mode: fs.ModeDir},
			"nested_dir/nested_subdir":            {Mode: fs.ModeDir},
			"nested_dir/nested_subdir/.gitignore": {Data: []byte("kept_dir\na.txt")},
			"nested_dir/nested_subdir/a.txt":      {Data: []byte("content")},
			"kept_dir":                            {Mode: fs.ModeDir},
			"kept_dir/a.txt":                      {Data: []byte("content")},
		}
		walker := NewRepoWalkerFS(fsys)
		visited := make(map[string]bool)
		_ = walker.WalkRepo(func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			visited[path] = true
			return nil
		})

		for _, v := range [3]string{"ignored_dir", "ignored_dir/a.txt", "nested_dir/nested_subdir/a.txt"} {
			if visited[v] {
				t.Errorf("%s should have been skipped", v)
			}
		}

		for _, v := range [4]string{"nested_dir", "nested_dir/nested_subdir", "kept_dir", "kept_dir/a.txt"} {
			if !visited[v] {
				t.Errorf("%s should have been visited", v)
			}
		}
	})
}
