//go:build integration
// +build integration

package medusa

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type dummyFileInfo struct{}

func (d dummyFileInfo) IsDir() bool        { return false }
func (d dummyFileInfo) ModTime() time.Time { return time.Now() }
func (d dummyFileInfo) Mode() fs.FileMode  { return 0755 }
func (d dummyFileInfo) Name() string       { return "1" }
func (d dummyFileInfo) Size() int64        { return 100 }
func (d dummyFileInfo) Sys() any           { return nil }

func generateSource(source string) error {
	files := make([]File, 0)
	for i := 0; i < 10; i++ {
		files = append(files, File{
			Path:     "posts/category/something/" + strconv.Itoa(i) + ".md",
			FileInfo: dummyFileInfo{},
			content:  []byte("# Header!"),
		})
	}
	err := writeFiles(source, files)
	if err != nil {
		return err
	}

	return nil
}

func noExistingDirs(t *testing.T, source string, destination string) {
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source dir exist, can't proceed with test")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatalf("destination dir exist, can't proceed with test")
	}
}
func removeDirs(dirs ...string) {
	for _, dir := range dirs {
		err := os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
	}
}

func compareFileState(t *testing.T, files []File, destination string) {
	for _, file := range files {
		if _, err := os.Stat(destination); os.IsNotExist(err) {
			t.Fatalf("file %v not present in %v. it should be", file.Path, destination)
		}

		filePath := filepath.Join(destination, file.Path)

		osFile, err := os.Open(filePath)
		if err != nil {
			panic(err) // this is unexpected
		}

		fileBytes, err := io.ReadAll(osFile)
		if err != nil {
			panic(err) // this is unexpected
		}

		if !bytes.Equal(fileBytes, file.content) {
			t.Fatalf("file contents %v does not match. \"%v\" != \"%v\"",
				file.Path,
				string(fileBytes),
				string(file.content),
			)
		}

	}
}

func TestPassthrough(t *testing.T) {
	source := "./src"
	destination := "./build"

	noExistingDirs(t, source, destination)

	err := generateSource(source)
	if err != nil {
		panic(err)
	}

	b := NewBuilder()
	b.Source(source)
	b.Destination(destination)
	b.Use(func(f *[]File, s *Store) error {
		var fileStoreIndex int
		for i, file := range *f {
			if strings.HasSuffix(file.Path, "0.md") {
				fileStoreIndex = i
				(*f)[i].SetContent([]byte("This is the first file."))
				(*f)[i].Store["test"] = "exists"
			}
		}
		(*s)["test"] = fileStoreIndex
		return nil
	})
	b.Use(func() Transformer {
		return func(f *[]File, s *Store) error {
			i, ok := (*s)["test"].(int)
			if !ok {
				t.Fatalf("store value not passed on correctly")
			}

			value, ok := (*f)[i].Store["test"]
			if !ok || value != "exists" {
				t.Fatalf("store value not passed on correctly")
			}
			return nil
		}

	}())

	err = b.Build()

	if err != nil {
		t.Fatalf("err != nil: %v", err)
	}
	compareFileState(t, b.files, destination)
	removeDirs(source, destination)
}
