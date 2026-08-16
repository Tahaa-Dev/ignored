package ignored

import (
	"testing"
)

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"Simple", "*.log", false},
		{"Directory", "build/", false},
		{"Negative", "!important.txt", false},
		{"Empty", "", true},
		{"Comment", "# this is a comment", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ParsePattern(tt.pattern)
			if (p.Err() != nil) != tt.wantErr {
				t.Errorf("ParsePattern() error = %v, wantErr %v", p.Err(), tt.wantErr)
			}
		})
	}
}

func TestPattern_Match(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		isDir   bool
		rootDir string
		want    bool
	}{
		{"Match Simple", "*.log", "/root/app.log", false, "/root", true},
		{"No Match Simple", "*.log", "/root/app.txt", false, "/root", false},
		{"Match Directory", "bin/", "/root/bin", true, "/root", true},
		{"Match Negative", "!tmp.log", "/root/tmp.log", false, "/root", false},
		{"Match Negative True", "!tmp.log", "/root/other.log", false, "/root", true},
		{"Nested Directory", "node_modules/", "/root/a/node_modules", true, "/root", true},
		{"Literal Pattern", "/config.json", "/root/config.json", false, "/root", true},
		{"Glob Star", "src/*.go", "/root/src/main.go", false, "/root", true},
		{"Double Star Deep", "**/test.txt", "/root/a/b/c/test.txt", false, "/root", true},
		{"Double Star Start", "**/a.go", "/root/s/a.go", false, "/root", true},
		{"Double Star Middle", "a/**/b.go", "/root/a/c/d/b.go", false, "/root", true},
		{
			"Ignore Directory Content",
			"node_modules/",
			"/root/node_modules/app/index.js",
			false,
			"/root",
			true,
		},
		{
			"Ignore Directory Content Dir Or File",
			"node_modules",
			"/root/node_modules/app/index.js",
			false,
			"/root",
			true,
		},
		{
			"Ignore Directory Content Anchored",
			"/node_modules/app/",
			"/root/node_modules/app/index.js",
			false,
			"/root",
			true,
		},
		{
			"Ignore Directory Content Anchored Dir Or File",
			"/node_modules/app",
			"/root/node_modules/app/index.js",
			false,
			"/root",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ParsePattern(tt.pattern)
			if p.Err() != nil {
				t.Fatalf("Failed to parse pattern %s: %v", tt.pattern, p.Err())
			}
			got, err := p.Match(tt.path, tt.isDir, tt.rootDir)
			if err != nil {
				t.Errorf("Match() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}
