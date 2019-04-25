// Ensure all the text file ends with a newline character at the end of file (EOF)
// https://stackoverflow.com/questions/729692/why-should-text-files-end-with-a-newline
package sanitychecks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	includeFilePrefixes = []string{"Makefile", "Dockerfile", "Gopkg"}

	includeFileSuffixes = []string{".go", ".yaml", ".md", ".sh", ".mk"}

	// relative to root directory
	includePaths = []string{".editorconfig", ".gitignore", "OWNERS", "make/gofmt_exclude"}

	// relative to root directory
	excludePaths = []string{"vendor", "build/_output", ".git", ".vendor-new"}
)

func TestNewlineAtEOF(t *testing.T) {
	curdir, err := os.Getwd()
	require.NoError(t, err)
	rootdir := filepath.Dir(filepath.Dir(curdir))
	excludePathsMap := map[string]bool{}
	for _, v := range excludePaths {

		excludePathsMap[filepath.Join(rootdir, v)] = true
	}
	includePathsMap := map[string]bool{}
	for _, v := range includePaths {

		includePathsMap[filepath.Join(rootdir, v)] = true
	}

	errorFiles := []string{}
	err = filepath.Walk(rootdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if _, ok := excludePathsMap[path]; ok {
			return filepath.SkipDir
		}
		filename := filepath.Base(path)
		if !info.IsDir() {
			proceed := false
			for _, v := range includeFilePrefixes {
				if strings.HasPrefix(filename, v) {
					proceed = true
					continue
				}
			}
			for _, v := range includeFileSuffixes {
				if strings.HasSuffix(filename, v) {
					proceed = true
					continue
				}
			}
			if _, ok := includePathsMap[path]; ok {
				proceed = true
			}
			if proceed {

				file, err := os.Open(path)
				if err != nil {
					panic(err)
				}
				defer file.Close()

				buf := make([]byte, 1)
				stat, err := os.Stat(path)
				if err != nil {
					t.Error("cannot stat file:", path)
				}
				start := stat.Size() - 1
				_, err = file.ReadAt(buf, start)
				if err != nil {
					t.Error("cannot read file:", path)
				} else {
					if '\n' != buf[0] {
						errorFiles = append(errorFiles, path)
					}

				}
			} else {
				t.Log("not checking for newline at EOF:", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Error("error walking the path:", err)
		return
	}
	if len(errorFiles) > 0 {
		t.Error("files without newline at the EOF:")
		for _, v := range errorFiles {
			t.Log(v)
		}
	}
}
