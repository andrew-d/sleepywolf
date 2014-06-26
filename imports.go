package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getImportPath(inputPath string) (string, error) {
	p, err := filepath.Abs(inputPath)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(p)
	gopaths := strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator))

	for _, path := range gopaths {
		gpath, err := filepath.Abs(path)
		if err != nil {
			continue
		}

		// Try to make a relative path from this GOPATH entry to our dir
		// If we can, this is a valid import path, and we can use it.
		rel, err := filepath.Rel(filepath.ToSlash(gpath), dir)
		if err != nil {
			return "", err
		}

		// The path should start with "src/", or the Go tool won't pick it up.
		if len(rel) < 4 || rel[:4] != "src"+string(os.PathSeparator) {
			continue
		}

		// Strip the leading "src/"
		return rel[4:], nil
	}

	return "", fmt.Errorf("Could not find source directory: GOPATH=%q REL=%q", gopaths, dir)
}
