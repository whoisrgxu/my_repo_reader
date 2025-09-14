# Repository Context

## File System Location

/Users/Roger/Documents/PersonalProject/myreporeader

## Git Info

- Commit: 4d04210e3cb4a39a2d5ee6291c344de49d244a1b
- Branch: main
- Author: Roger
- Date: Sat Sep 13 22:00:33 2025 -0400

## Structure

```
README.md
go.mod
internal/
  filters/
    ignore.go
    textdetect.go
main.go
myreporeader
output.md
```
## File Contents

### File: README.md
```md
# myreporeader

A small CLI that prints a human‑readable snapshot of a codebase. It shows the directory tree, selected file contents, Git metadata, and a summary of file/line counts — while respecting `.gitignore` (including nested ones) and common build/cache directories.

---

## Features

- **Structure view**: Directory tree that hides dotfiles (except `.gitignore`) and skips ignored paths.
- **File contents**: Inlines text files (or only a specific extension via `--include`) with fenced code blocks.
- **Smart ignoring**: Loads every `.gitignore` under the target path and applies rules from the file’s directory up to the repo root. Also includes sensible defaults (e.g., `node_modules/`, `.next/`, `dist/`, `__pycache__/`, etc.).
- **Accurate summary**: Counts only text files; if inside a Git repo, counts only Git‑tracked files (via `git ls-files`). Falls back to an ignore‑aware filesystem walk when Git is not available.
- **Binary detection**: Heuristic detection to avoid printing or counting binary artifacts and large bundles.

---

## Installation

```bash
# From the project root
go build -o myreporeader

# Optional: move onto your PATH
# Linux/macOS (adjust path as needed)
sudo mv myreporeader /usr/local/bin/
```
**Requirements:** Go 1.20+

---

## Usage

```text
myreporeader <path> [--include .ext] [o outputfile]
```

### Arguments

- `<path>`  
  File or directory to read.

- `--include .ext`  
  Only include files with the given extension in the **File Contents** section (summary still respects ignore and text detection).

- `o outputfile`  
  Write Markdown output to `outputfile` instead of stdout.

### Examples

```bash
# Print a whole repo to stdout
myreporeader .

# Write a Markdown snapshot
myreporeader ./my-app o output.md

# Only include JS files in the File Contents section
myreporeader ./my-app --include .js o repo-js.md

# Target a single file
myreporeader ./src/app/page.js
```

---

## Output sections

- `# Repository Context`
  - **File System Location**
  - **Git Info** (Commit / Branch / Author / Date) — shown if the path is inside a Git repo
  - **Structure** — directory tree (respects ignore rules)
  - **File Contents** — inlined text files; optionally filtered by `--include .ext`
  - **Summary** — total text files and total lines counted

---

## How ignoring works

The tool recursively loads `.gitignore` files from every directory under the target path.

For a given file, patterns from its own directory’s `.gitignore` are applied first, then parent directories up to the root you passed in.

Supported rule types (pragmatic subset of `.gitignore`):

- Directory rules ending with `/` (e.g., `node_modules/`, `build/`) match the directory itself and everything underneath it.
- Root‑anchored rules starting with `/` (e.g., `/dist`, `/build/`) are matched from the repository root.
- Extension rules (e.g., `*.log`).
- Plain names matched anywhere in the path (e.g., `coverage`).

Default ignore patterns are also applied for common ecosystems (Node, Python, Java, .NET, Go, Rust, etc.). See `internal/filters/filters.go`.

> **Note:** Negations (`!pattern`) and `**` recursive globs are not currently supported.

---

## What counts as a text file

Text detection is implemented in `internal/filters`:

1. **Extension hint:** A broad allow‑list of common source/markup/config extensions (e.g., `js/ts/tsx/go/py/java/cpp/json/yaml/toml/css/html/md`, and many others).
2. **Sniffing:** Reads the first ~8 KB; if a NUL byte is found, it is considered binary. If the sample is valid UTF‑8 (or ASCII), it is considered text.
3. **Empty files** are considered text.

This keeps binary blobs (WASM, images, compiled artifacts, large `.map` files, etc.) out of both **File Contents** and **Summary**.

---

## Git‑aware line counting

When the target folder is inside a Git repo, `myreporeader` runs:

```bash
git -C <root> ls-files -z
```

It then counts lines only in those tracked files (still filtered by ignore rules and text detection).

If Git is not available, it falls back to an ignore‑aware filesystem walk.

---

## Errors and exit status

- Non‑fatal issues (e.g., unreadable files) are logged to stderr and skipped.
- The program returns `0` on success; fatal errors (e.g., invalid path) will exit non‑zero.

---

## Project layout

```text
.
├── internal/
│   └── filters/
│       ├── filters.go          # IsTextFile, MatchPattern, DefaultIgnorePatterns
│       └── text_ext.go         # Extension allow‑list
├── main.go                     # CLI entry
└── README.md
```

---

## Limitations / TODO

- No support for `.gitignore` negations (`!pattern`) or `**` recursive globs.
- Language detection for code fences is extension‑based.
- Large repositories may produce large outputs; consider `--include .ext` to focus.

---

## License

MIT

```
### File: go.mod
```mod
module github.com/whoisrgxu/myreporeader

go 1.25.1

```
### File: internal/filters/ignore.go
```go
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

```
### File: internal/filters/textdetect.go
```go
package filters

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Broad text/code extensions
var TextExt = map[string]struct{}{
	// docs & markup
	".txt": {}, ".md": {}, ".mdx": {}, ".rst": {}, ".adoc": {}, ".asciidoc": {},
	".tex": {}, ".bib": {}, ".org": {}, ".textile": {},

	// data / logs
	".csv": {}, ".tsv": {}, ".psv": {}, ".ndjson": {}, ".log": {}, ".properties": {},

	// config / serialization
	".json": {}, ".json5": {}, ".jsonc": {},
	".yaml": {}, ".yml": {}, ".toml": {}, ".ini": {}, ".cfg": {}, ".conf": {}, ".env": {},

	// html/xml/svg
	".html": {}, ".htm": {}, ".xhtml": {}, ".xml": {}, ".xsd": {}, ".xsl": {}, ".xslt": {}, ".dtd": {}, ".svg": {},

	// styles
	".css": {}, ".scss": {}, ".sass": {}, ".less": {}, ".styl": {},

	// web templates
	".ejs": {}, ".pug": {}, ".jade": {}, ".hbs": {}, ".mustache": {}, ".njk": {}, ".twig": {}, ".liquid": {},

	// js/ts ecosystem
	".js": {}, ".mjs": {}, ".cjs": {}, ".jsx": {},
	".ts": {}, ".tsx": {},
	".vue": {}, ".svelte": {}, ".astro": {},

	// go
	".go": {}, ".tmpl": {}, ".mod": {}, ".sum": {},

	// python
	".py": {}, ".pyi": {}, ".pyw": {}, ".pyx": {}, ".pxd": {}, ".pxi": {},

	// ruby
	".rb": {}, ".erb": {}, ".rake": {}, ".gemspec": {},

	// php
	".php": {}, ".phtml": {}, ".php3": {}, ".php4": {}, ".php5": {}, ".php7": {}, ".php8": {},

	// java / groovy / kotlin / scala
	".java": {}, ".jsp": {}, ".groovy": {}, ".gradle": {}, ".gvy": {}, ".gy": {}, ".gsh": {},
	".kt": {}, ".kts": {}, ".ktm": {},
	".scala": {}, ".sc": {}, ".sbt": {},

	// c / c++ / objc / swift
	".c": {}, ".h": {}, ".hpp": {}, ".hh": {}, ".hxx": {}, ".cpp": {}, ".cc": {}, ".cxx": {}, ".ino": {}, ".ipp": {},
	".m": {}, ".mm": {}, ".pch": {},
	".swift": {}, ".xcconfig": {}, ".pbxproj": {}, ".xcscheme": {}, ".xcworkspacedata": {}, ".plist": {}, ".strings": {},

	// .NET / F#
	".cs": {}, ".csx": {}, ".fs": {}, ".fsi": {}, ".fsx": {},

	// rust
	".rs": {}, ".ron": {},

	// haskell / ocaml
	".hs": {}, ".lhs": {}, ".cabal": {},
	".ml": {}, ".mli": {}, ".re": {}, ".rei": {},

	// erlang / elixir
	".erl": {}, ".hrl": {}, ".ex": {}, ".exs": {}, ".eex": {}, ".leex": {}, ".heex": {},

	// lua
	".lua": {}, ".rockspec": {},

	// shells
	".sh": {}, ".bash": {}, ".zsh": {}, ".ksh": {}, ".fish": {}, ".command": {},

	// powershell / batch
	".ps1": {}, ".psm1": {}, ".psd1": {}, ".bat": {}, ".cmd": {},

	// build / tooling
	".cmake": {}, ".ninja": {}, ".bazel": {}, ".bzl": {},

	// infra / IaC
	".tf": {}, ".tfvars": {}, ".hcl": {}, ".cue": {}, ".dhall": {},

	// idl / schema
	".proto": {}, ".thrift": {}, ".avdl": {},

	// query / graph
	".sql": {}, ".psql": {}, ".mysql": {}, ".cql": {}, ".graphql": {}, ".gql": {},

	// diagrams
	".plantuml": {}, ".puml": {}, ".dot": {}, ".gv": {}, ".mermaid": {}, ".mmd": {},

	// data science
	".r": {}, ".R": {}, ".Rmd": {}, ".qmd": {}, ".jl": {},
}

// Well-known text filenames (no extension)
var TextFilenames = map[string]struct{}{
	"Makefile": {}, "CMakeLists.txt": {},
	"Dockerfile": {}, ".dockerignore": {},
	".gitignore": {}, ".gitattributes": {}, ".gitmodules": {},
	".npmrc": {}, ".nvmrc": {}, ".prettierrc": {}, ".eslintignore": {}, ".eslintrc": {},
	"SConstruct": {}, "SConscript": {},
	"BUILD": {}, "BUILD.bazel": {}, "WORKSPACE": {}, "WORKSPACE.bazel": {},
	"Gemfile": {}, "Rakefile": {}, "Vagrantfile": {}, "Procfile": {}, "Jenkinsfile": {},
	"LICENSE": {}, "LICENSE.md": {}, "COPYING": {}, "README": {}, "README.md": {},
	"CHANGELOG": {}, "CHANGELOG.md": {}, "NOTICE": {}, "AUTHORS": {},
}

func hasTextyName(path string) bool {
	base := filepath.Base(path)
	if _, ok := TextFilenames[base]; ok {
		return true
	}
	ext := strings.ToLower(filepath.Ext(base))
	_, ok := TextExt[ext]
	return ok
}

// Robust content sniff
func isProbablyTextFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	const sniff = 8192
	buf := make([]byte, sniff)
	n, _ := f.Read(buf)
	if n == 0 {
		return true // empty counts as text
	}
	s := buf[:n]

	// NUL byte → binary
	if bytes.IndexByte(s, 0x00) != -1 {
		return false
	}

	// MIME hint
	mime := http.DetectContentType(s)
	if strings.HasPrefix(mime, "text/") {
		return true
	}
	switch mime {
	case "application/json", "application/javascript", "application/xml", "image/svg+xml":
		return true
	}

	// UTF-8 → text
	if utf8.Valid(s) {
		return true
	}

	// Printable ASCII heuristic
	printable := 0
	for _, b := range s {
		if b == 9 || b == 10 || b == 13 || (b >= 32 && b <= 126) {
			printable++
		}
	}
	return float64(printable)/float64(len(s)) >= 0.95
}

// Exported helper used by main
func IsTextFile(path string) bool {
	return hasTextyName(path) || isProbablyTextFile(path)
}

```
### File: main.go
```go
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"

	filters "github.com/whoisrgxu/myreporeader/internal/filters"
)

type Directory struct {
	ParentPath string
	Name       string
	Indent     string
}

type GitInfo struct {
	Hash   string
	Branch string
	Author string
	Date   string
}

// Per-directory .gitignore rules
var gitignoreRules = map[string][]string{}

// ---------------- .gitignore handling ----------------

func loadGitignores(root string) {
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			gitignorePath := filepath.Join(path, ".gitignore")
			data, err := os.ReadFile(gitignorePath)
			if err == nil {
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					gitignoreRules[path] = append(gitignoreRules[path], line)
				}
			}
		}
		return nil
	})
}

// Check ignore using .gitignore (walking up to root) + default patterns.
func isIgnored(path string, root string) bool {
	abs, _ := filepath.Abs(path)
	abs = filepath.Clean(abs)

	// 1) .gitignore rules from the file's dir up to root
	dir := filepath.Dir(abs)
	for {
		patterns := gitignoreRules[dir]
		relFromDir, _ := filepath.Rel(dir, abs)
		relFromDir = filepath.ToSlash(relFromDir)

		for _, pat := range patterns {
			pat = strings.TrimSpace(pat)
			if pat == "" || strings.HasPrefix(pat, "#") {
				continue
			}
			if filters.MatchPattern(relFromDir, pat) {
				return true
			}
		}

		if dir == root {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// 2) Default cross-ecosystem patterns relative to repo root
	relFromRoot, _ := filepath.Rel(root, abs)
	relFromRoot = filepath.ToSlash(relFromRoot)
	for _, pat := range filters.DefaultIgnorePatterns {
		if filters.MatchPattern(relFromRoot, pat) {
			return true
		}
	}

	return false
}

// ---------------- Git helpers (for accurate summary) ----------------

func isGitRepo(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

func listGitTrackedFiles(root string) ([]string, error) {
	cmd := exec.Command("git", "-C", root, "ls-files", "-z")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(out, []byte{0})
	files := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		files = append(files, filepath.Join(root, string(p)))
	}
	return files, nil
}

func countFilesAndLinesGit(root string) (int, int, error) {
	files, err := listGitTrackedFiles(root)
	if err != nil {
		return 0, 0, err
	}

	fileCount := 0
	lineCount := 0

	for _, f := range files {
		if isIgnored(f, root) {
			continue
		}
		if !filters.IsTextFile(f) {
			continue
		}
		lines, err := countLinesInFile(f)
		if err != nil {
			continue
		}
		fileCount++
		lineCount += lines
	}
	return fileCount, lineCount, nil
}

// ---------------- Core FS helpers ----------------

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (d Directory) getPath() string {
	return filepath.Join(d.ParentPath, d.Name)
}

func (d Directory) readEntries() []os.DirEntry {
	path := d.getPath()
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	return entries
}

// Robust line counter (handles long lines)
func countLinesInFile(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	count := 0
	for {
		_, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func countFilesAndLines(paths []string, root string) (int, int) {
	fileCount := 0
	lineCount := 0

	for _, path := range paths {
		if isIgnored(path, root) {
			continue
		}

		if isDir(path) {
			entries, err := os.ReadDir(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading dir %s: %v\n", path, err)
				continue
			}

			for _, entry := range entries {
				// Hide dotfiles except .gitignore
				if strings.HasPrefix(entry.Name(), ".") && entry.Name() != ".gitignore" {
					continue
				}
				childPath := filepath.Join(path, entry.Name())
				if isIgnored(childPath, root) {
					continue
				}

				cf, cl := countFilesAndLines([]string{childPath}, root)
				fileCount += cf
				lineCount += cl
			}
		} else {
			if !filters.IsTextFile(path) {
				continue
			}
			lines, err := countLinesInFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error counting lines in %s: %v\n", path, err)
				continue
			}
			fileCount++
			lineCount += lines
		}
	}
	return fileCount, lineCount
}

func getNonHiddenEntries(entries []os.DirEntry) []os.DirEntry {
	var result []os.DirEntry
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") && e.Name() != ".gitignore" {
			continue
		}
		result = append(result, e)
	}
	return result
}

// ---------------- Printing ----------------

func (d Directory) printStructure(w io.Writer, root string) {
	path := d.getPath()
	entries := getNonHiddenEntries(d.readEntries())

	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		if isIgnored(childPath, root) {
			continue
		}

		if entry.IsDir() {
			fmt.Fprint(w, d.Indent, entry.Name(), "/\n")
			childDir := Directory{
				ParentPath: path,
				Name:       entry.Name(),
				Indent:     d.Indent + "  ",
			}
			childDir.printStructure(w, root)
		} else {
			fmt.Fprint(w, d.Indent, entry.Name(), "\n")
		}
	}
}

func (d Directory) identifyFileType(entry os.DirEntry) string {
	ext := filepath.Ext(entry.Name())
	if len(ext) > 0 {
		return ext[1:]
	}
	return ""
}

func (d Directory) printFiles(entries []os.DirEntry, rootPath string, w io.Writer, skipFile string, include string, root string) {
	entries = getNonHiddenEntries(entries)

	for _, entry := range entries {
		fullPath := filepath.Join(d.getPath(), entry.Name())
		if isIgnored(fullPath, root) {
			continue
		}

		if entry.IsDir() {
			childDir := Directory{
				ParentPath: d.getPath(),
				Name:       entry.Name(),
				Indent:     d.Indent + "  ",
			}
			childDir.printFiles(childDir.readEntries(), rootPath, w, skipFile, include, root)
			continue
		}

		if include != "" && filepath.Ext(entry.Name()) != include {
			continue
		}

		absFull, _ := filepath.Abs(fullPath)
		absSkip, _ := filepath.Abs(skipFile)
		if skipFile != "" && absFull == absSkip {
			continue
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Fprintf(w, "Error reading %s: %v\n", fullPath, err)
			continue
		}

		// Only print text-ish files
		if utf8.Valid(data) && filters.IsTextFile(fullPath) {
			relPath, err := filepath.Rel(rootPath, fullPath)
			if err != nil {
				relPath = fullPath
			}
			fileType := d.identifyFileType(entry)
			fmt.Fprintf(w, "### File: %v\n", relPath)
			fmt.Fprintf(w, "```%v\n", fileType)
			fmt.Fprintf(w, "%v\n```\n", string(data))
		}
	}
}

// ---------------- Git info ----------------

func (d Directory) GetLatestCommit() (*GitInfo, error) {
	cmd := exec.Command("git", "-C", d.ParentPath, "log", "-1", "--pretty=format:%H|%an|%ad")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	parts := strings.SplitN(out.String(), "|", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected git log format")
	}

	branchCmd := exec.Command("git", "-C", d.ParentPath, "rev-parse", "--abbrev-ref", "HEAD")
	var branchOut bytes.Buffer
	branchCmd.Stdout = &branchOut
	if err := branchCmd.Run(); err != nil {
		return nil, err
	}

	return &GitInfo{
		Hash:   parts[0],
		Author: parts[1],
		Date:   parts[2],
		Branch: strings.TrimSpace(branchOut.String()),
	}, nil
}

// ---------------- Main output ----------------

func output(args []string) {
	length := len(args)
	var folderPath string
	var w io.Writer
	var include string
	var skipFile string
	var filePaths []string

	targetPath, err := filepath.Abs(args[1])
	if err != nil {
		panic(err)
	}

	if isDir(targetPath) {
		folderPath = targetPath
		filePaths = nil
		loadGitignores(folderPath)
	} else {
		folderPath = filepath.Dir(targetPath)
		filePaths = []string{targetPath}
		loadGitignores(folderPath)
	}

	dir := Directory{
		ParentPath: folderPath,
		Name:       "",
		Indent:     "",
	}

	if length > 2 && args[length-2] == "o" {
		ww, err := os.Create(args[length-1])
		if err != nil {
			panic(err)
		}
		w = ww
		absSkip, _ := filepath.Abs(args[length-1])
		skipFile = absSkip
	} else {
		w = os.Stdout
		skipFile = ""
	}

	if len(args) > 2 && args[2] == "--include" {
		include = filepath.Ext(args[3])
	} else {
		include = ""
	}

	fmt.Fprintf(w, "# Repository Context\n\n")
	fmt.Fprintf(w, "## File System Location\n\n")
	fmt.Fprintln(w, folderPath)
	fmt.Fprintf(w, "## Git Info\n\n")

	gitInfo, err := dir.GetLatestCommit()
	if err == nil {
		fmt.Fprintf(w, "- Commit: %v\n", gitInfo.Hash)
		fmt.Fprintf(w, "- Branch: %v\n", gitInfo.Branch)
		fmt.Fprintf(w, "- Author: %v\n", gitInfo.Author)
		fmt.Fprintf(w, "- Date: %v\n", gitInfo.Date)
	}

	fmt.Fprintf(w, "## Structure\n\n")
	fmt.Fprintln(w, "```")
	dir.printStructure(w, folderPath)
	fmt.Fprintln(w, "```")

	fmt.Fprintf(w, "## File Contents\n\n")
	if len(filePaths) == 0 {
		dir.printFiles(dir.readEntries(), folderPath, w, skipFile, include, folderPath)
	} else {
		for _, filePath := range filePaths {
			if isIgnored(filePath, folderPath) {
				continue
			}
			data, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(w, "Error reading %s: %v\n", filePath, err)
				continue
			}
			if utf8.Valid(data) && filters.IsTextFile(filePath) {
				fileType := strings.TrimPrefix(filepath.Ext(filePath), ".")
				fmt.Fprintf(w, "### File: %v\n", filepath.Base(filePath))
				fmt.Fprintf(w, "```%v\n", fileType)
				fmt.Fprintf(w, "%v\n```\n", string(data))
			}
		}
	}

	// Summary (prefer Git-tracked; fallback to FS walk)
	var fileCount, lineCount int
	if len(filePaths) == 0 {
		if isGitRepo(folderPath) {
			if fc, lc, err := countFilesAndLinesGit(folderPath); err == nil {
				fileCount, lineCount = fc, lc
			} else {
				entries := getNonHiddenEntries(dir.readEntries())
				var childPaths []string
				for _, entry := range entries {
					childPath := filepath.Join(folderPath, entry.Name())
					if isIgnored(childPath, folderPath) {
						continue
					}
					childPaths = append(childPaths, childPath)
				}
				fileCount, lineCount = countFilesAndLines(childPaths, folderPath)
			}
		} else {
			entries := getNonHiddenEntries(dir.readEntries())
			var childPaths []string
			for _, entry := range entries {
				childPath := filepath.Join(folderPath, entry.Name())
				if isIgnored(childPath, folderPath) {
					continue
				}
				childPaths = append(childPaths, childPath)
			}
			fileCount, lineCount = countFilesAndLines(childPaths, folderPath)
		}
	} else {
		fileCount, lineCount = countFilesAndLines(filePaths, folderPath)
	}

	fmt.Fprintf(w, "## Summary\n- Total files: %v\n- Total lines: %v\n", fileCount, lineCount)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: myreporeader <path> [--include .ext] [o outputfile]")
		return
	}
	output(os.Args)
}

```
## Summary
- Total files: 5
- Total lines: 910
