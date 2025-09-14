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
