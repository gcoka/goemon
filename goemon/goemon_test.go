package goemon_test

import (
	"os"
	"testing"

	"github.com/gcoka/goemon/goemon"
)

func Test_ListTarget(t *testing.T) {
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

	type args struct {
		watches []string
		ignores []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"ignore vendor", args{[]string{"."}, []string{".git", ".git/**", "vendor"}}, []string{".env", "Makefile", "README.md", "cmd", "cmd/somecmd", "cmd/somecmd/root.go", "hello", "hello/hello.go", "main.go"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watches, ignores := goemon.CompileGlobs(tt.args.watches), goemon.CompileGlobs(tt.args.ignores)
			got := listMapKeys(goemon.ListTarget(watches, ignores))

			if !deepEqualSorted(got, tt.want) {
				t.Errorf("ListTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeExt(t *testing.T) {
	type args struct {
		ext []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{"1", args{[]string{"go,md", "yml json"}}, []string{"go", "md", "yml", "json"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := goemon.NormalizeExt(tt.args.ext); !deepEqualSorted(got, tt.want) {
				t.Errorf("NormalizeExt() = %v, want %v", got, tt.want)
			}
		})
	}
}
