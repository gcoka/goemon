package goemon

import (
	"os"
	"reflect"
	"testing"
)

func TestGlobWalker_Walk(t *testing.T) {

	var root string
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir("..")

	root, _ = os.Getwd()
	tests := []struct {
		name     string
		patterns []string
		want     []string
	}{
		// TODO: Add test cases.
		{"example/*.go", []string{"example/*.go"}, []string{"example/test.go"}},
		{"Makefile", []string{"Makefile"}, []string{"Makefile"}},
		{"cmd/* & example/", []string{"cmd/*", "example"}, []string{"cmd/root.go", "example"}},
		{".editorconfig", []string{".editorconfig"}, []string{".editorconfig"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gw := &GlobWalker{
				globs: CompileGlobs(tt.patterns),
				root:  root,
			}
			if got := gw.Walk("."); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GlobWalker.Walk() = %v, want %v", got, tt.want)
			}
		})
	}
}
