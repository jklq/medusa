package markdown

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"git.sr.ht/~relay/medusa"
	"github.com/yuin/goldmark"
)

func New() medusa.Transformer {
	return func(files *[]medusa.File, store *medusa.Store) error {
		var buf bytes.Buffer
		for i := range *files {
			file := &((*files)[i])
			if filepath.Ext(file.Path) != ".md" {
				continue
			}
			if err := goldmark.Convert(file.Content(), &buf); err != nil {
				return err
			}
			content, err := io.ReadAll(&buf)
			if err != nil {
				return err
			}

			file.SetContent(content)
			file.Path = strings.TrimSuffix(file.Path, ".md") + ".html"
		}
		return nil
	}
}
