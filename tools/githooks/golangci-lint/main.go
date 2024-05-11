//go:build githooks
// +build githooks

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

var golangciLintArgs = []string{"run", "github.com/golangci/golangci-lint/cmd/golangci-lint", "run"}

func main() {
	/*
		Problem:
			1) pre-commit tool runs hook and pass staged files to it
			2) golangci-lint takes files, but cannot run some
				linters without info from entire package -> so, lint fails

		Solution:
			get package names from staged files and run golangci-lint on packages
	*/
	fileNames := os.Args[1:]
	golangciLintArgs = append(golangciLintArgs, getPkgNames(fileNames)...)
	fmt.Printf("DEBUG %v\n", golangciLintArgs)
	out, err := exec.Command("go", golangciLintArgs...).CombinedOutput()
	fmt.Printf("%s\n", out)
	if err != nil {
		os.Exit(1)
	}
}

func getPkgNames(files []string) []string {
	uniqDirNames := make(map[string]struct{})
	for _, name := range files {
		uniqDirNames[path.Dir(name)] = struct{}{}
	}
	dirNames := make([]string, 0, len(uniqDirNames))
	for dir := range uniqDirNames {
		dirNames = append(dirNames, dir)
	}
	return dirNames
}
