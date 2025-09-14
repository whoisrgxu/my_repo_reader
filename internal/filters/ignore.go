package filters

import (
	"path/filepath"
	"strings"
)

// Cross-ecosystem default ignore patterns
var DefaultIgnorePatterns = []string{
	// Node.js
	"node_modules/", "package-lock.json", "yarn.lock", "pnpm-lock.yaml",
	".next/", "dist/", "build/", "coverage/",

	// Python
	"__pycache__/", ".venv/", ".mypy_cache/", ".pytest_cache/",
	"Pipfile.lock", "poetry.lock",

	// Java
	"target/", "build/", ".gradle/", "*.iml",

	// .NET / C#
	"bin/", "obj/", "packages/",

	// Go
	"vendor/", "*.exe", "*.out",

	// Rust
	"target/", "Cargo.lock",

	// General
	".DS_Store", "Thumbs.db",
}

// MatchPattern: simplified .gitignore-like matcher.
//
// Supports:
//   - directory rules like "node_modules/" (match at root or ANY subdir)
//   - anchored rules like "/node_modules" or "/build/"
//   - extension rules like "*.log"
//   - plain names like "dist" (match in any subdir)
func MatchPattern(rel, pattern string) bool {
	rel = filepath.ToSlash(rel)

	anchored := strings.HasPrefix(pattern, "/")
	p := pattern
	if anchored {
		p = p[1:]
	}
	p = filepath.ToSlash(p)

	// Directory rule (ends with "/")
	if strings.HasSuffix(p, "/") {
		dir := strings.TrimSuffix(p, "/")
		if dir == "" {
			return false
		}
		if anchored {
			return rel == dir || strings.HasPrefix(rel, dir+"/")
		}
		// unanchored: match anywhere in the path
		return rel == dir ||
			strings.HasSuffix(rel, "/"+dir) ||
			strings.HasPrefix(rel, dir+"/") ||
			strings.Contains(rel, "/"+dir+"/")
	}

	// Extension rule: "*.ext"
	if strings.HasPrefix(p, "*.") {
		return strings.HasSuffix(rel, p[1:])
	}

	// Anchored plain rule
	if anchored {
		return rel == p || strings.HasPrefix(rel, p+"/")
	}

	// Unanchored plain rule: match anywhere
	return rel == p || strings.HasSuffix(rel, "/"+p) || strings.Contains(rel, "/"+p+"/")
}
