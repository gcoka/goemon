package goemon

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gobwas/glob"
)

// GlobWalker is globbing match file walker.
type GlobWalker struct {
	globs []glob.Glob
	root  string
}

// CompileGlobs compiles pattern strings into Glob.
func CompileGlobs(patterns []string) []glob.Glob {
	globs := make([]glob.Glob, 0, len(patterns))
	for _, p := range patterns {
		var g glob.Glob
		cp := filepath.Clean(p)
		if cp == "." || cp == ".." {
			g = glob.MustCompile("**", '/', filepath.Separator)
		} else {
			g = glob.MustCompile(p, '/', filepath.Separator)
		}
		globs = append(globs, g)
	}
	return globs
}

// NewGlobWalker initialize GlobWalker
func NewGlobWalker(g []glob.Glob) *GlobWalker {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}

	return &GlobWalker{
		g,
		absRoot,
	}
}

func (gw *GlobWalker) isTarget(path string, info os.FileInfo) bool {
	absPath, _ := filepath.Abs(path)
	rel, _ := filepath.Rel(gw.root, absPath)

	for _, v := range gw.globs {
		if v.Match(rel) {
			return true
		}
		if v.Match(filepath.Base(rel)) {
			return true
		}
	}
	return false
}

// Walk finds all files which matches the glob pattern.
func (gw *GlobWalker) Walk(path string, walkFn filepath.WalkFunc) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.Mode().IsRegular() {
		if gw.isTarget(path, fi) {
			err := walkFn(path, fi, nil)
			if err != nil {
				return err
			}
		}
		return nil
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		p := filepath.Join(path, file.Name())

		if gw.isTarget(p, file) {
			walkFn(p, file, nil)
		}

		if file.IsDir() {
			err := gw.Walk(p, walkFn)
			if err != nil {
				return err
			}
		}

	}

	return nil
}
