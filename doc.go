// Package ignored provides a way to handle gitignore and similar ignore files (e.g. dockerignore) at the pattern level, file level and repository level through 3 main interfaces:
//
// ## Pattern
//
// The [Pattern] interface provides methods to handle ignore patterns individually which is parsed from a string via the [ParsePattern] function.
//
// ### Example
//
//	import (
//		"fmt"
//		"github.com/Tahaa-Dev/ignored"
//	)
//
//	func IsIgnored(pattern string, path string, isDir bool, rootDir string) bool {
//		pat := ignored.ParsePattern(pattern)
//		isIgnored, err := pat.Match(path, isDir, rootDir)
//		if err != nil {
//			fmt.Println("Error while matching path:", err.Error())
//		}
//	}
//
// ## Matcher
//
// The [Matcher] interface provides methods to handle multiple ignore patterns at once which is constructed via the [NewMatcher] function.
//
// ### Example
//
//	import "github.com/Tahaa-Dev/ignored"
//
//	var rootDir = "/home/myproject"
//	func IsIgnored(matcher ignored.Matcher, path string, isDir bool, ignorePath string) bool {
//		if ignorePath != "" {
//			matcher.ExtendFromFile(ignorePath)
//		}
//		return matcher.Match(path, isDir)
//	}
//
// ## RepoWalker
//
// The [RepoWalker] interface provides an ignore file compliant wrapper for [fs.WalkDir] which is constructed via the [NewRepoWalker] function.
//
// ### Example
//
//	import (
//		"fmt"
//		"os"
//		"github.com/Tahaa-Dev/ignored"
//	)
//
//	func PrintRepo(repoWalker ignored.RepoWalker) {
//		repoWalker.WalkRepo(func(path string, d fs.DirEntry, err error) error {
//			suffix := ""
//			if d.IsDir() {
//				suffix = "/"
//			}
//			fmt.Printf("%s%s\n", path, suffix)
//		})
//		if repoWalker.Err() != nil {
//			fmt.Fprintf(os.Stderr, "Error while printing dir: %s\n", repoWalker.Err().Error())
//		}
//	}
package ignored
