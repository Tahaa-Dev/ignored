package ignored

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNewMatcher(t *testing.T) {
	m := NewMatcher("/root/path", "*.log", "build/")
	if m == nil {
		t.Fatal("NewMatcher returned nil")
	}

	impl, ok := m.(*matcher)
	if !ok {
		t.Fatal("NewMatcher did not return a *matcher instance")
	}

	expectedRoot := filepath.ToSlash("/root/path")
	if impl.rootDir != expectedRoot {
		t.Errorf("Expected rootDir to be %q, got %q", expectedRoot, impl.rootDir)
	}

	if len(impl.patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(impl.patterns))
	}

	if m.Len() != 2 {
		t.Errorf("Expected Len() to be 2, got %d", m.Len())
	}

	if m.Err() != nil {
		t.Errorf("Expected no errors, got %v", m.Err())
	}
}

func TestMatcher_SetRootDir(t *testing.T) {
	m := NewMatcher("/root/path")
	m.SetRootDir("/new/root/path/")

	impl := m.(*matcher)
	expectedRoot := filepath.ToSlash("/new/root/path/")
	if impl.rootDir != expectedRoot {
		t.Errorf("Expected rootDir to be %q, got %q", expectedRoot, impl.rootDir)
	}
}

func TestMatcher_Extend(t *testing.T) {
	m := NewMatcher("/root")
	m.Extend("*.go", "*.txt")

	if m.Len() != 2 {
		t.Errorf("Expected Len() to be 2, got %d", m.Len())
	}

	m.Extend("", "# comment")
	if m.Err() != nil {
		t.Errorf(
			"Expected error when extending with empty/comment patterns to be nil, but got: %s",
			m.Err().Error(),
		)
	}

	if m.Len() != 2 {
		t.Errorf("Expected 2 valid patterns, got %d", m.Len())
	}
}

func TestMatcher_ExtendFromPatterns(t *testing.T) {
	m := NewMatcher("/root", "*.log")

	p1 := ParsePattern("*.go")
	p2 := ParsePattern("*.txt")

	m.ExtendFromPatterns(p1, p2)

	if m.Len() != 3 {
		t.Errorf("Expected Len() to be 3, got %d", m.Len())
	}

	impl := m.(*matcher)
	// Verify that patterns are indeed the ones we appended and haven't been overwritten
	// The order should be: *.log, *.go, *.txt
	if impl.patterns[0].Err() != nil || impl.patterns[1] != p1 || impl.patterns[2] != p2 {
		t.Errorf(
			"Patterns in matcher are not in correct order or were overwritten. Patterns: %v",
			impl.patterns,
		)
	}
}

type errReader struct{}

func (errReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestMatcher_ExtendFromReader(t *testing.T) {
	t.Run("Valid Reader", func(t *testing.T) {
		m := NewMatcher("/root")
		buf := bytes.NewBufferString("*.go\n# comment line\n*.txt\n")
		m.ExtendFromReader(buf)

		if m.Len() != 2 {
			t.Errorf("Expected Len() to be 2, got %d", m.Len())
		}
		if m.Err() != nil {
			t.Errorf("Expected Err() to be nil, got %v", m.Err())
		}
	})

	t.Run("Reader with Error", func(t *testing.T) {
		m := NewMatcher("/root")
		m.ExtendFromReader(errReader{})

		if m.Err() == nil {
			t.Error("Expected error from errReader, got nil")
		}
	})
}

func TestMatcher_ExtendFromFile(t *testing.T) {
	t.Run("Valid File", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "ignore-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()

		content := "*.go\n*.json\n"
		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		defer func() { _ = tmpFile.Close() }()

		m := NewMatcher("/root")
		m.ExtendFromFile(tmpFile.Name())

		if m.Len() != 2 {
			t.Errorf("Expected Len() to be 2, got %d", m.Len())
		}
		if m.Err() != nil {
			t.Errorf("Expected Err() to be nil, got %v", m.Err())
		}
	})

	t.Run("Missing File", func(t *testing.T) {
		m := NewMatcher("/root")
		m.ExtendFromFile("non-existent-file-xyz")

		if m.Err() == nil {
			t.Error("Expected error when loading from non-existent file, got nil")
		}
	})
}

func TestMatcher_Match(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		isDir    bool
		want     bool
	}{
		{"Match first pattern", []string{"*.log", "*.go"}, "/root/app.log", false, true},
		{"Match second pattern", []string{"*.log", "*.go"}, "/root/main.go", false, true},
		{"Match none", []string{"*.log", "*.go"}, "/root/main.py", false, false},
		{"Empty patterns", []string{}, "/root/main.py", false, false},
		{"Match directory", []string{"build/"}, "/root/build", true, true},
		{
			"Match directory with file inside",
			[]string{"build/"},
			"/root/build/main.js",
			false,
			true,
		},
		{
			"Negative override (whitelist)",
			[]string{"*.log", "!important.log"},
			"/root/app.log",
			false,
			true,
		},
		{
			"Negative override (whitelist) exact",
			[]string{"*.log", "!important.log"},
			"/root/important.log",
			false,
			false,
		},
		{
			"Negative override order significance",
			[]string{"!important.log", "*.log"},
			"/root/important.log",
			false,
			true,
		},
		{"Anchored root match", []string{"/bin/"}, "/root/bin/app", false, true},
		{"Anchored root mismatch deep", []string{"/bin/"}, "/root/src/bin/app", false, false},
		{"Unanchored match deep", []string{"bin/"}, "/root/src/bin/app", false, true},
		{
			"Glob double star",
			[]string{"**/logs/**/*.log"},
			"/root/src/logs/app/app.log",
			false,
			true,
		},
		{
			"Glob double star mismatch",
			[]string{"**/logs/**/*.log"},
			"/root/src/notlogs/app.log",
			false,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher("/root", tt.patterns...)
			if got := m.Match(tt.path, tt.isDir); got != tt.want {
				t.Errorf("Match(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.want)
			}
		})
	}
}

func TestMatcher_ChainedExtend(t *testing.T) {
	m := NewMatcher("/root").
		Extend("*.go").
		Extend("*.txt").
		Extend("*.log")

	if m.Len() != 3 {
		t.Errorf("Expected Len() to be 3, got %d", m.Len())
	}

	// Verify chained extends match as expected
	if !m.Match("/root/main.go", false) {
		t.Error("Expected main.go to be matched")
	}
	if !m.Match("/root/notes.txt", false) {
		t.Error("Expected notes.txt to be matched")
	}
	if !m.Match("/root/app.log", false) {
		t.Error("Expected app.log to be matched")
	}
}

func TestMatcher_removePatterns(t *testing.T) {
	m := NewMatcher("/root", "*.go", "*.txt", "*.log")
	impl := m.(*matcher)

	impl.removePatterns(2)
	if m.Len() != 2 {
		t.Errorf("Expected Len() to be 2, got %d", m.Len())
	}

	if impl.patterns[0].Err() != nil || impl.patterns[1].Err() != nil {
		t.Error("Patterns changed unexpectedly after removePatterns")
	}
}
