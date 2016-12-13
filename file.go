package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func getFile(filePath string) (io.Reader, error) {
	file, err := os.Open(filePath)
	return file, err
}

func walk(path string, f func(filePath string)) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.ToLower(filepath.Ext(info.Name())) == ".go" {
			f(path)
		}
		return nil
	})
}

func findDirectory(goRoot, goPath, srcDirectory, pkg string) (string, error) {
	path := filepath.SplitList(srcDirectory)
	path = append(path, "", "vendor")
	count := len(path)
	dest := ""
	for count > 1 {
		path[count-2] = path[count-1]
		path = path[:count-1]
		vendor := filepath.Join(path...)
		pkgPath := filepath.Join(vendor, pkg)
		if pkgDir(pkgPath) == nil {
			dest = pkgPath
			break
		}
	}
	if dest == "" {
		pkgPath := filepath.Join(goPath, pkg)
		if pkgDir(pkgPath) == nil {
			dest = pkgPath
		}
	}
	if dest == "" {
		pkgPath := filepath.Join(goRoot, pkg)
		if pkgDir(pkgPath) == nil {
			dest = pkgPath
		}
	}
	if dest == "" {
		return "", fmt.Errorf("can't find pkg dir: %s", pkg)
	}
	return "", nil
}

func pkgDir(pkgPath string) error {
	info, err := os.Stat(pkgPath)
	if err == nil && info.IsDir() {
		return nil
	}
	return err
}
