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

Download the latest release from [GitHub Releases](https://github.com/whoisrgxu/my_repo_reader/releases), 
unpack the archive for your OS/arch, and place `myreporeader` in your `$PATH`.

Example (Linux/macOS):

```bash
curl -L https://github.com/whoisrgxu/my_repo_reader/releases/latest/download/myreporeader -o /usr/local/bin/myreporeader
chmod +x /usr/local/bin/myreporeader
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
