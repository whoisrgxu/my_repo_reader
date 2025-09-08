package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Directory struct {
	ParentPath string
	Name       string
	indent     string
	// SpecificExt string
	// Children []Directory
	// File     []string
}

type GitInfo struct {
	Hash   string
	Branch string
	Author string
	Date   string
}

func (d Directory) getPath() string {
	return d.ParentPath + "/" + d.Name
}
func (d Directory) readEntries() []os.DirEntry {

	path := d.getPath()
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	return entries
}
func (d Directory) printSpecificExtension() {
	// parentPath := d.ParentPath
	// path := parentPath + "/" + d.Name
	// entries, err := os.ReadDir(path)
	// if err != nil {
	// 	panic(err)
	// }
	// for _, entry := range entries {
	// 	if entry.IsDir() {
	// 		fmt.Printf("%s/%s\n", path, entry.Name())
	// 		childDir := Directory{
	// 			ParentPath: parentPath + "/" + d.Name,
	// 			Name:       entry.Name(),
	// 		}
	// 		childDir.PrintStructure()
	// 	}
	// }
}
func (d Directory) printStructure() {

	path := d.getPath()
	entries := d.readEntries()
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("%s%s/\n", d.indent, entry.Name())
			childDir := Directory{
				ParentPath: path,
				Name:       entry.Name(),
				indent:     d.indent + "  ",
			}
			childDir.printStructure()
		} else {
			fmt.Printf("%s%s\n", d.indent, entry.Name())
		}
	}
}

func (d Directory) GetLatestCommit() (*GitInfo, error) {
	// commit hash | author | date | message
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

	// get current branch
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

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: myreporeader <command>")
		return
	} else if len(os.Args) >= 2 {
		var targetPath string
		cmd := os.Args[1]
		if cmd == "." {
			currentPath, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			targetPath = currentPath
		} else {
			if strings.Contains(cmd, ".") {
				{ /* to implement */
				}
			} else {
				targetPath = cmd
			}
		}
		fmt.Printf("# Repository Context\n\n")
		fmt.Printf("## File System Location\n\n")
		fmt.Println(targetPath)
		fmt.Printf("## Git Info\n\n")

		dir := Directory{
			ParentPath: targetPath,
			Name:       "",
			indent:     "",
		}

		fmt.Println(dir.GetLatestCommit())
		fmt.Printf("## Structure\n\n")
		fmt.Println("```")
		dir.printStructure()
		fmt.Println("```")
	}
}
