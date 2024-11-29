package layouts

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	"git.sr.ht/~relay/medusa"
)

var (
	ErrNoLayoutPattern  = errors.New("no layout pattern specified")
	ErrNoContentPattern = errors.New("no content pattern specified")
	ErrNoLayouts        = errors.New("no layouts in specified patterns")
)

type ErrInvalidLayoutName struct{ path string }

func (e ErrInvalidLayoutName) Error() string {
	return fmt.Sprintf("invalid layout value at: %v", e.path)
}

type ErrLayoutNotFound struct {
	layout string
	path   string
}

func (e ErrLayoutNotFound) Error() string {
	return fmt.Sprintf("did not find layout \"%v\" as defined in: %v", e.layout, e.path)
}

type Config struct {
	LayoutPatterns  []string
	ContentPatterns []string
}

func fileMatchesPatterns(patterns []string, file medusa.File) (bool, error) {
	matchOne := false
	for _, pattern := range patterns {
		match, err := filepath.Match(pattern, file.Path)
		if err != nil {
			return false, err
		}
		if match {
			matchOne = true
		}
	}
	return matchOne, nil
}

// The data accessible from the template
type TemplateData struct {
	// Local file struct
	File medusa.File

	// Global store
	Global medusa.Store

	Content template.HTML
}

// Creates a the layout transformer
func New(cfg Config) medusa.Transformer {
	return func(files *[]medusa.File, store *medusa.Store) error {
		if len(cfg.LayoutPatterns) == 0 {
			return ErrNoLayoutPattern
		}
		if len(cfg.ContentPatterns) == 0 {
			return ErrNoContentPattern
		}

		var layouts = map[string]*template.Template{}
		var i = 0

		var defaultLayout *template.Template
		var layoutDefaultFound bool
		var selectedLayout *template.Template
		for i < len(*files) {
			file := &((*files)[i])

			match, err := fileMatchesPatterns(cfg.LayoutPatterns, *file)
			if err != nil {
				return err
			}
			if !match {
				i++
				continue
			}

			layoutName := file.Path

			selectedLayout, err = template.New(layoutName).Parse(string(file.Content()))
			if err != nil {
				return err
			}

			if strings.HasPrefix(file.FileInfo.Name(), "default") {
				layoutDefaultFound = true
				defaultLayout = selectedLayout
			}

			layouts[layoutName] = selectedLayout

			(*files)[i] = (*files)[len(*files)-1]
			(*files) = (*files)[:len(*files)-1]
		}
		if !layoutDefaultFound {
			defaultLayout = selectedLayout
		}

		if len(layouts) == 0 {
			return ErrNoLayouts
		}
		for i := range *files {
			file := &((*files)[i])

			match, err := fileMatchesPatterns(cfg.ContentPatterns, *file)
			if err != nil {
				return err
			}
			if !match {
				i++
				continue
			}

			layout := defaultLayout

			if name, ok := file.Frontmatter["layout"]; ok {
				layoutName, ok := name.(string)
				if !ok {
					return ErrInvalidLayoutName{file.Path}
				}

				layout, ok = layouts[layoutName]
				if !ok {
					return ErrLayoutNotFound{layoutName, file.Path}
				}
			}

			var templateData TemplateData

			templateData.File = *file
			templateData.Global = *store
			templateData.Content = template.HTML(file.Content())

			var newContentBuffer bytes.Buffer

			err = layout.Execute(&newContentBuffer, templateData)

			if err != nil {
				return err
			}

			newContent, err := io.ReadAll(&newContentBuffer)

			if err != nil {
				return err
			}

			file.SetContent(newContent)

		}

		return nil
	}
}
