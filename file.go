package medusa

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/adrg/frontmatter"
)

type ErrFrontmatter struct {
	path string
}

func (e ErrFrontmatter) Error() string {
	return fmt.Sprintf("failed to parse frontmatter at: %v", e.path)
}

// future max size to load into memory?
// or maybe a user specified option?
// or based on free memory?
// const maxInMemorySize = 100 * 1024 * 1024 // 100MB

type File struct {
	Path     string
	FileInfo fs.FileInfo
	Store    Store

	// the yaml/toml/json frontmatter of the file
	Frontmatter Store

	content []byte
}

// Get the contents of the file
func (f *File) Content() []byte {
	return f.content
}

// Set the contents of the file
func (f *File) SetContent(bytes []byte) {
	f.content = bytes
}

func (b *Builder) srcWalker(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.IsDir() {
		return nil
	}

	fileinfo, err := d.Info()
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var fm = make(map[string]any)
	var content []byte

	if !b.skipFrontmatter {
		content, err = frontmatter.Parse(file, &fm)
		if err != nil {
			b.log.Error("failed to parse frontmatter", "file", path)
			return ErrFrontmatter{path: path}
		}
	} else {
		content, err = io.ReadAll(file)
		if err != nil {
			return err
		}
	}

	sourceRelPath, err := filepath.Rel(b.source, path)
	if err != nil {
		return err
	}

	b.files = append(b.files, File{
		FileInfo:    fileinfo,
		Path:        sourceRelPath,
		Store:       make(Store),
		Frontmatter: fm,

		content: content,
	})
	return nil
}
