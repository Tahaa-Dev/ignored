# ignored

`ignored` is a Go package designed to handle Gitignore-like ignore patterns at
three levels:

- **Pattern level:** Parsing and matching individual ignore rules.
- **Matcher level:** Managing a collection of ignore patterns.
- **Repository Walker level:** An `fs.WalkDir` wrapper that automatically
  applies ignore rules while traversing a directory.

---

## Core Interfaces

### Pattern

The `Pattern` interface handles individual ignore rules. Patterns are parsed
from strings, supporting negation (e.g., `!file.txt`) and directory-specific
rules.

### Matcher

The `Matcher` interface manages a collection of patterns. It allows dynamic
addition of patterns from strings or files (like `.gitignore`), providing
efficient matching for paths.

### RepoWalker

The `RepoWalker` provides a wrapper around `fs.WalkDir`, making it easy to
traverse a filesystem while respecting ignore rules dynamically loaded from
ignore files (e.g., `.gitignore`) in each directory.

---

## Installation

```bash
go get github.com/Tahaa-Dev/ignored
```

---

## Quick Examples

### Using RepoWalker

```go
import (
    "fmt"
    "io/fs"
    "github.com/Tahaa-Dev/ignored"
)

func main() {
    walker := ignored.NewRepoWalker("/path/to/project").SetIgnoreFileName(".dockerignore")
    
    err := walker.WalkRepo(func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        fmt.Println("Visiting:", path)
        return nil
    })
    
    if err != nil {
        // handle fs.WalkDir error
    }
    if walker.Err() != nil {
        // Handle accumulated errors
    }
}
```

### Using Matcher

```go
import (
    "fmt"
    "github.com/Tahaa-Dev/ignored"
)

func main() {
    matcher := ignored.NewMatcher("/path/to/project/root", "*.log", "node_modules/")
    
    if matcher.Match("node_modules/pkg/main.js", false) {
        fmt.Println("Path is ignored")
    }
}
```

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file
for details.

---

## Development

- See [CONTRIBUTING.md](CONTRIBUTING.md) for information about contributing to
  the project.
- See [CHANGELOG.md](CHANGELOG.md) for news and changes about the project.
