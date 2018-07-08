package goemon_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func setup(t *testing.T) string {
	tmpdir, err := ioutil.TempDir("", "goemon_test")
	if err != nil {
		t.Fatal(err)
	}

	// Assume golang cli package development.
	os.MkdirAll(filepath.Join(tmpdir, ".git"), 0755)
	ioutil.WriteFile(filepath.Join(tmpdir, ".git/config"), []byte{}, 0644)

	os.MkdirAll(filepath.Join(tmpdir, "vendor/github.com/somepkg-go/.github"), 0755)
	ioutil.WriteFile(filepath.Join(tmpdir, "vendor/github.com/somepkg-go/README.md"), []byte{}, 0644)
	ioutil.WriteFile(filepath.Join(tmpdir, "vendor/github.com/somepkg-go/Makefile"), []byte{}, 0644)
	ioutil.WriteFile(filepath.Join(tmpdir, "vendor/github.com/somepkg-go/main.go"), []byte{}, 0644)

	os.MkdirAll(filepath.Join(tmpdir, "cmd/somecmd"), 0755)
	ioutil.WriteFile(filepath.Join(tmpdir, "cmd/somecmd/root.go"), []byte{}, 0644)

	ioutil.WriteFile(filepath.Join(tmpdir, "main.go"), []byte{}, 0644)
	ioutil.WriteFile(filepath.Join(tmpdir, "README.md"), []byte{}, 0644)
	ioutil.WriteFile(filepath.Join(tmpdir, "Makefile"), []byte{}, 0644)
	ioutil.WriteFile(filepath.Join(tmpdir, ".env"), []byte{}, 0644)

	os.MkdirAll(filepath.Join(tmpdir, "hello"), 0755)
	ioutil.WriteFile(filepath.Join(tmpdir, "hello/hello.go"), []byte{}, 0644)

	return tmpdir
}

func deepEqualSorted(got []string, want []string) bool {
	sort.Strings(got)
	sort.Strings(want)
	return reflect.DeepEqual(got, want)
}

func listMapKeys(m map[string]os.FileInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
