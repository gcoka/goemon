package goemon_test

import (
	"os"
	"testing"

	"github.com/gcoka/goemon/goemon"
)

func TestGlobWalker_Walk(t *testing.T) {

	tmpDir := setup(t)
	defer os.RemoveAll(tmpDir)

	cDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cDir)

	tests := []struct {
		name     string
		patterns []string
		want     []string
	}{
		{"*.go", []string{"*.go"}, []string{"main.go", "cmd/somecmd/root.go", "hello/hello.go", "vendor/github.com/somepkg-go/main.go"}},
		{"Makefile", []string{"Makefile"}, []string{"Makefile", "vendor/github.com/somepkg-go/Makefile"}},
		{"cmd/* hello", []string{"cmd/*", "hello"}, []string{"cmd/somecmd", "hello"}},
		{".env", []string{".env"}, []string{".env"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gw := goemon.NewGlobWalker(goemon.CompileGlobs(tt.patterns))
			gotFiles := make([]string, 0)
			err := gw.Walk(".", func(p string, fi os.FileInfo, e error) error {
				gotFiles = append(gotFiles, p)
				return nil
			})
			if err != nil {
				t.Errorf("GlobWalker.Walk() returns error = %v", err)
			}
			if !deepEqualSorted(gotFiles, tt.want) {
				t.Errorf("GlobWalker.Walk() = %v, want %v", gotFiles, tt.want)
			}
		})
	}
}
