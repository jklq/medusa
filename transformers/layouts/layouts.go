package layouts

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"git.sr.ht/~relay/medusa"
)

var (
	ErrNoLayoutPattern  = errors.New("no layout pattern specified")
	ErrNoContentPattern = errors.New("no content pattern specified")
	ErrNoLayouts        = errors.New("no layouts found matching specified patterns")
	ErrNoDefaultLayout  = errors.New("no default layout found and no layouts available to choose from")
)

type ErrInvalidLayoutName struct{ path string }

func (e ErrInvalidLayoutName) Error() string {
	return fmt.Sprintf("invalid layout value (must be a string) in frontmatter for: %v", e.path)
}

type ErrLayoutNotFound struct {
	layout string
	path   string
}

func (e ErrLayoutNotFound) Error() string {
	return fmt.Sprintf("layout \"%v\" not found (referenced in: %v)", e.layout, e.path)
}

type Config struct {
	// Glob patterns to identify layout and partial files.
	LayoutPatterns []string
	// Glob patterns to identify content files that need layout application.
	ContentPatterns []string
}

func fileMatchesPatterns(patterns []string, file medusa.File) (bool, error) {
	matchOne := false
	for _, pattern := range patterns {
		match, err := filepath.Match(pattern, file.Path)
		if err != nil {
			return false, fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		}
		if match {
			matchOne = true
			break // Found a match, no need to check further patterns
		}
	}
	return matchOne, nil
}

// TemplateData is the data structure accessible from within the templates.
type TemplateData struct {
	// File holds metadata and frontmatter for the current content file.
	File medusa.File

	// Global provides access to the shared data store.
	Global medusa.Store

	// Content is the rendered HTML content of the current file.
	// It's marked as template.HTML to prevent double-escaping.
	Content template.HTML
}

func New(cfg Config) medusa.Transformer {
	return func(files *[]medusa.File, store *medusa.Store) error {
		if len(cfg.LayoutPatterns) == 0 {
			return ErrNoLayoutPattern
		}
		if len(cfg.ContentPatterns) == 0 {
			return ErrNoContentPattern
		}

		masterTmpl := template.New("")
		var layoutFiles []medusa.File
		var contentFiles []medusa.File
		var defaultLayoutName string
		var lastLayoutName string

		// Pass 1: Separate files and parse layouts
		remainingFiles := (*files)[:0]
		for _, file := range *files {
			isLayout, err := fileMatchesPatterns(cfg.LayoutPatterns, file)
			if err != nil {
				return err
			}

			if isLayout {
				layoutFiles = append(layoutFiles, file)
				layoutName := file.Path

				_, err := masterTmpl.New(layoutName).Parse(string(file.Content()))
				if err != nil {
					return fmt.Errorf("failed to parse layout/partial '%s': %w", layoutName, err)
				}

				if strings.HasPrefix(filepath.Base(file.Path), "default.") {
					if defaultLayoutName == "" {
						defaultLayoutName = layoutName
					}
				}
				lastLayoutName = layoutName
			} else {
				isContent, err := fileMatchesPatterns(cfg.ContentPatterns, file)
				if err != nil {
					return err
				}

				if isContent {
					contentFiles = append(contentFiles, file)
				} else {
					remainingFiles = append(remainingFiles, file)
				}
			}
		}
		*files = remainingFiles

		if len(layoutFiles) == 0 {
			needsLayout := false
			for _, contentFile := range contentFiles {
				_, hasLayoutKey := contentFile.Frontmatter["layout"]
				if len(contentFiles) > 0 || hasLayoutKey {
					needsLayout = true
					break
				}
			}
			if needsLayout {
				return ErrNoLayouts
			}
			*files = append(*files, contentFiles...)
			return nil
		}

		if defaultLayoutName == "" {
			if lastLayoutName == "" {
				return ErrNoDefaultLayout
			}
			defaultLayoutName = lastLayoutName
		}

		// Pass 2: Process content files
		processedContentFiles := make([]medusa.File, 0, len(contentFiles))
		for _, file := range contentFiles {
			targetLayoutName := defaultLayoutName

			if name, ok := file.Frontmatter["layout"]; ok {
				layoutNameStr, ok := name.(string)
				if !ok {
					return ErrInvalidLayoutName{path: file.Path}
				}
				targetLayoutName = layoutNameStr
			}

			if masterTmpl.Lookup(targetLayoutName) == nil {
				if targetLayoutName == defaultLayoutName {
					return fmt.Errorf("default layout '%s' (required by '%s') not found or failed to parse", defaultLayoutName, file.Path)
				}
				return ErrLayoutNotFound{layout: targetLayoutName, path: file.Path}
			}

			templateData := TemplateData{
				File:    file,
				Global:  *store,
				Content: template.HTML(file.Content()),
			}

			var newContentBuffer bytes.Buffer
			err := masterTmpl.ExecuteTemplate(&newContentBuffer, targetLayoutName, templateData)
			if err != nil {
				return fmt.Errorf("failed to execute layout '%s' for file '%s': %w", targetLayoutName, file.Path, err)
			}

			file.SetContent(newContentBuffer.Bytes())
			processedContentFiles = append(processedContentFiles, file)
		}

		*files = append(*files, processedContentFiles...)

		return nil
	}
}
